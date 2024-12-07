package helpers

import (
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/commands/git_commands"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/jesseduffield/lazygit/pkg/utils"
	"github.com/rivo/uniseg"
	"github.com/samber/lo"
)

type ICommitsHelper interface {
	UpdateCommitPanelView(message string)
}

type CommitsHelper struct {
	c *HelperCommon

	getCommitSummary              func() string
	setCommitSummary              func(string)
	getCommitDescription          func() string
	getUnwrappedCommitDescription func() string
	setCommitDescription          func(string)
}

var _ ICommitsHelper = &CommitsHelper{}

func NewCommitsHelper(
	c *HelperCommon,
	getCommitSummary func() string,
	setCommitSummary func(string),
	getCommitDescription func() string,
	getUnwrappedCommitDescription func() string,
	setCommitDescription func(string),
) *CommitsHelper {
	return &CommitsHelper{
		c:                             c,
		getCommitSummary:              getCommitSummary,
		setCommitSummary:              setCommitSummary,
		getCommitDescription:          getCommitDescription,
		getUnwrappedCommitDescription: getUnwrappedCommitDescription,
		setCommitDescription:          setCommitDescription,
	}
}

func (self *CommitsHelper) SplitCommitMessageAndDescription(message string) (string, string) {
	msg, description, _ := strings.Cut(message, "\n")
	return msg, strings.TrimSpace(description)
}

func (self *CommitsHelper) SetMessageAndDescriptionInView(message string) {
	summary, description := self.SplitCommitMessageAndDescription(message)

	self.setCommitSummary(summary)
	self.setCommitDescription(description)
	self.c.Contexts().CommitMessage.RenderCommitLength()
}

func (self *CommitsHelper) JoinCommitMessageAndUnwrappedDescription() string {
	if len(self.getUnwrappedCommitDescription()) == 0 {
		return self.getCommitSummary()
	}
	return self.getCommitSummary() + "\n" + self.getUnwrappedCommitDescription()
}

func TryRemoveHardLineBreaks(message string, autoWrapWidth int) string {
	messageRunes := []rune(message)
	lastHardLineStart := 0
	for i, r := range messageRunes {
		if r == '\n' {
			// Try to make this a soft linebreak by turning it into a space, and
			// checking whether it still wraps to the same result then.
			messageRunes[i] = ' '

			_, cursorMapping := gocui.AutoWrapContent(messageRunes[lastHardLineStart:], autoWrapWidth)

			// Look at the cursorMapping to check whether auto-wrapping inserted
			// a line break. If it did, there will be a cursorMapping entry with
			// Orig pointing to the position after the inserted line break.
			if len(cursorMapping) == 0 || cursorMapping[0].Orig != i-lastHardLineStart+1 {
				// It didn't, so change it back to a newline
				messageRunes[i] = '\n'
			}
			lastHardLineStart = i + 1
		}
	}

	return string(messageRunes)
}

func (self *CommitsHelper) SwitchToEditor() error {
	message := lo.Ternary(len(self.getCommitDescription()) == 0,
		self.getCommitSummary(),
		self.getCommitSummary()+"\n\n"+self.getCommitDescription())
	filepath := filepath.Join(self.c.OS().GetTempDir(), self.c.Git().RepoPaths.RepoName(), time.Now().Format("Jan _2 15.04.05.000000000")+".msg")
	err := self.c.OS().CreateFileWithContent(filepath, message)
	if err != nil {
		return err
	}

	self.CloseCommitMessagePanel()

	return self.c.Contexts().CommitMessage.SwitchToEditor(filepath)
}

func (self *CommitsHelper) UpdateCommitPanelView(message string) {
	if message != "" {
		self.SetMessageAndDescriptionInView(message)
		return
	}

	if self.c.Contexts().CommitMessage.GetPreserveMessage() {
		preservedMessage := self.c.Contexts().CommitMessage.GetPreservedMessageAndLogError()
		self.SetMessageAndDescriptionInView(preservedMessage)
		return
	}

	self.SetMessageAndDescriptionInView("")
}

type OpenCommitMessagePanelOpts struct {
	CommitIndex      int
	SummaryTitle     string
	DescriptionTitle string
	PreserveMessage  bool
	OnConfirm        func(summary string, description string) error
	OnSwitchToEditor func(string) error
	InitialMessage   string
}

func (self *CommitsHelper) OpenCommitMessagePanel(opts *OpenCommitMessagePanelOpts) {
	onConfirm := func(summary string, description string) error {
		self.CloseCommitMessagePanel()

		return opts.OnConfirm(summary, description)
	}

	self.c.Contexts().CommitMessage.SetPanelState(
		opts.CommitIndex,
		opts.SummaryTitle,
		opts.DescriptionTitle,
		opts.PreserveMessage,
		opts.InitialMessage,
		onConfirm,
		opts.OnSwitchToEditor,
	)

	self.UpdateCommitPanelView(opts.InitialMessage)

	self.c.Context().Push(self.c.Contexts().CommitMessage)
}

func (self *CommitsHelper) OnCommitSuccess() {
	// if we have a preserved message we want to clear it on success
	if self.c.Contexts().CommitMessage.GetPreserveMessage() {
		self.c.Contexts().CommitMessage.SetPreservedMessageAndLogError("")
	}
}

func (self *CommitsHelper) HandleCommitConfirm() error {
	summary, description := self.getCommitSummary(), self.getCommitDescription()

	if summary == "" {
		return errors.New(self.c.Tr.CommitWithoutMessageErr)
	}

	err := self.c.Contexts().CommitMessage.OnConfirm(summary, description)
	if err != nil {
		return err
	}

	return nil
}

func (self *CommitsHelper) CloseCommitMessagePanel() {
	if self.c.Contexts().CommitMessage.GetPreserveMessage() {
		message := self.JoinCommitMessageAndUnwrappedDescription()
		if message != self.c.Contexts().CommitMessage.GetInitialMessage() {
			self.c.Contexts().CommitMessage.SetPreservedMessageAndLogError(message)
		}
	} else {
		self.SetMessageAndDescriptionInView("")
	}

	self.c.Contexts().CommitMessage.SetHistoryMessage("")

	self.c.Views().CommitMessage.Visible = false
	self.c.Views().CommitDescription.Visible = false

	self.c.Context().Pop()
}

func (self *CommitsHelper) OpenCommitMenu(suggestionFunc func(string) []*types.Suggestion) error {
	var disabledReasonForOpenInEditor *types.DisabledReason
	if !self.c.Contexts().CommitMessage.CanSwitchToEditor() {
		disabledReasonForOpenInEditor = &types.DisabledReason{
			Text: self.c.Tr.CommandDoesNotSupportOpeningInEditor,
		}
	}

	menuItems := []*types.MenuItem{
		{
			Label: self.c.Tr.OpenInEditor,
			OnPress: func() error {
				return self.SwitchToEditor()
			},
			Key:            'e',
			DisabledReason: disabledReasonForOpenInEditor,
		},
		{
			Label: self.c.Tr.AddCoAuthor,
			OnPress: func() error {
				return self.addCoAuthor(suggestionFunc)
			},
			Key: 'c',
		},
		{
			Label: self.c.Tr.AddGitmoji,
			OnPress: func() error {
				return self.addGitmoji()
			},
			Key: 'g',
		},
		{
			Label: self.c.Tr.PasteCommitMessageFromClipboard,
			OnPress: func() error {
				return self.pasteCommitMessageFromClipboard()
			},
			Key: 'p',
		},
	}
	return self.c.Menu(types.CreateMenuOptions{
		Title: self.c.Tr.CommitMenuTitle,
		Items: menuItems,
	})
}

func (self *CommitsHelper) addCoAuthor(suggestionFunc func(string) []*types.Suggestion) error {
	self.c.Prompt(types.PromptOpts{
		Title:               self.c.Tr.AddCoAuthorPromptTitle,
		FindSuggestionsFunc: suggestionFunc,
		HandleConfirm: func(value string) error {
			commitDescription := self.getCommitDescription()
			commitDescription = git_commands.AddCoAuthorToDescription(commitDescription, value)
			self.setCommitDescription(commitDescription)
			return nil
		},
	})

	return nil
}

func (self *CommitsHelper) gitmojiSuggestions() func(string) []*types.Suggestion {
	mojis := []string{
		"➕ Add a dependency",
		"🧪 Add a failing test",
		"👷 Add or update CI build system",
		"🙈 Add or update a .gitignore file",
		"🥚 Add or update an easter egg",
		"📈 Add or update analytics or track code",
		"💫 Add or update animations and transitions",
		"🍱 Add or update assets",
		"👔 Add or update business logic",
		"🧵 Add or update code related to multithreading or concurrency",
		"🦺 Add or update code related to validation",
		"💡 Add or update comments in source code",
		"📦️ Add or update compiled files or packages",
		"🔧 Add or update configuration files",
		"👥 Add or update contributor(s)",
		"🔨 Add or update development scripts",
		"📝 Add or update documentation",
		"🩺 Add or update healthcheck",
		"📄 Add or update license",
		"🔊 Add or update logs",
		"🔐 Add or update secrets",
		"🌱 Add or update seed files",
		"📸 Add or update snapshots",
		"💬 Add or update text and literals",
		"💄 Add or update the UI and style files",
		"🏷️ Add or update types",
		"💸 Add sponsorships or money related infrastructure",
		"✅ Add, update, or pass tests",
		"🚩 Add, update, or remove feature flags",
		"🎉 Begin a project",
		"🥅 Catch errors",
		"🚑️ Critical hotfix",
		"🧐 Data exploration/inspection",
		"🚀 Deploy stuff",
		"🗑️ Deprecate code that needs to be cleaned up",
		"⬇️ Downgrade dependencies",
		"💚 Fix CI Build",
		"🐛 Fix a bug",
		"🚨 Fix compiler / linter warnings",
		"🔒️ Fix security or privacy issues",
		"✏️ Fix typos",
		"🔍️ Improve SEO",
		"♿️ Improve accessibility",
		"🧑‍💻 Improve developer experience",
		"⚡️ Improve performance",
		"🎨 Improve structure / format of the code",
		"🚸 Improve user experience / usability",
		"🧱 Infrastructure related changes",
		"🌐 Internationalization and localization",
		"💥 Introduce breaking changes",
		"✨ Introduce new features",
		"🏗️ Make architectural changes",
		"🔀 Merge branches",
		"🤡 Mock things",
		"🚚 Move or rename resources (e.g.: files, paths, routes)",
		"🗃️ Perform database related changes",
		"⚗️ Perform experiments",
		"📌 Pin dependencies to specific versions",
		"♻️ Refactor code",
		"🔖 Release / Version tags",
		"➖ Remove a dependency",
		"🔥 Remove code or files",
		"⚰️ Remove dead code",
		"🔇 Remove logs",
		"⏪️ Revert changes",
		"🩹 Simple fix for a non-critical issue",
		"👽️ Update code due to external API changes",
		"⬆️ Upgrade dependencies",
		"🚧 Work in progress",
		"🛂 Work on code related to authorization, roles and permissions",
		"📱 Work on responsive design",
		"💩 Write bad code that needs to be improved",
		"🍻 Write code drunkenly",
	}
	return func(input string) []*types.Suggestion {
		var matches []string
		if input == "" {
			matches = mojis
		} else {
			matches = utils.FilterStrings(input, mojis, true)
		}
		return matchesToSuggestions(matches)
	}
}

// isEmoji checks if the rune belongs to known emoji Unicode ranges
func isEmoji(value string) bool {
	if len(value) > 1 {
		if uniseg.GraphemeClusterCount(value) == 1 {
			return true
		}
	}
	r := rune(value[0])
	// Emoji ranges defined by Unicode
	// Emoticons (0x1F600-0x1F64F)
	if r >= 0x1F600 && r <= 0x1F64F {
		return true
	}
	// Miscellaneous Symbols and Pictographs (0x1F300-0x1F5FF)
	if r >= 0x1F300 && r <= 0x1F5FF {
		return true
	}
	// Transport and Map Symbols (0x1F680-0x1F6FF)
	if r >= 0x1F680 && r <= 0x1F6FF {
		return true
	}
	// Supplemental Symbols and Pictographs (0x1F900-0x1F9FF)
	if r >= 0x1F900 && r <= 0x1F9FF {
		return true
	}
	// Miscellaneous Symbols (0x2600-0x26FF)
	if r >= 0x2600 && r <= 0x26FF {
		return true
	}
	// Dingbats (0x2700-0x27BF)
	if r >= 0x2700 && r <= 0x27BF {
		return true
	}
	// Other known emoji-related ranges could be added here.
	// For example, flags, family emojis, and emoji modifiers are also valid.

	// Check if the character is a regional indicator letter (flags)
	if r >= 0x1F1E6 && r <= 0x1F1FF {
		return true
	}

	// Check if the character is part of the variation selector (VS) range (emoji modifiers)
	if r >= 0xFE00 && r <= 0xFE0F {
		return true
	}

	// Some other characters like emoji modifiers might also qualify as part of an emoji
	// Return false if the character is not part of any known emoji block
	return false
}

func (self *CommitsHelper) addGitmoji() error {
	self.c.Prompt(types.PromptOpts{
		Title:               self.c.Tr.AddGitmojiPromptTitle,
		FindSuggestionsFunc: self.gitmojiSuggestions(),
		HandleConfirm: func(value string) error {
			if len(value) == 0 {
				return nil
			}
			commitDescription := self.getCommitSummary()
			currentGitmoji, rest, _, _ := uniseg.FirstGraphemeClusterInString(commitDescription, -1)
			gitmoji, _, _, _ := uniseg.FirstGraphemeClusterInString(value, -1)
			// If no emoji was directly selected, choose the first suggestion
			if !isEmoji(gitmoji) {
				suggestions := self.gitmojiSuggestions()(value)
				gitmoji, _, _, _ = uniseg.FirstGraphemeClusterInString(suggestions[0].Label, -1)
			}
			if len(currentGitmoji) > 0 && isEmoji(currentGitmoji) {
				commitDescription = gitmoji + rest
			} else {
				commitDescription = gitmoji + commitDescription + currentGitmoji
			}
			self.setCommitSummary(commitDescription)
			return nil
		},
	})

	return nil
}

func (self *CommitsHelper) pasteCommitMessageFromClipboard() error {
	message, err := self.c.OS().PasteFromClipboard()
	if err != nil {
		return err
	}
	if message == "" {
		return nil
	}

	if currentMessage := self.JoinCommitMessageAndUnwrappedDescription(); currentMessage == "" {
		self.SetMessageAndDescriptionInView(message)
		return nil
	}

	// Confirm before overwriting the commit message
	self.c.Confirm(types.ConfirmOpts{
		Title:  self.c.Tr.PasteCommitMessageFromClipboard,
		Prompt: self.c.Tr.SurePasteCommitMessage,
		HandleConfirm: func() error {
			self.SetMessageAndDescriptionInView(message)
			return nil
		},
	})

	return nil
}
