package utils

import (
	"github.com/rivo/uniseg"
	"github.com/samber/lo"
)

type Gitmoji struct {
	Emoji string
	Label string
}

var gitmojis = []Gitmoji{
	{Emoji: "➕", Label: "Add a dependency"},
	{Emoji: "🧪", Label: "Add a failing test"},
	{Emoji: "👷", Label: "Add or update CI build system"},
	{Emoji: "🙈", Label: "Add or update a .gitignore file"},
	{Emoji: "🥚", Label: "Add or update an easter egg"},
	{Emoji: "📈", Label: "Add or update analytics or track code"},
	{Emoji: "💫", Label: "Add or update animations and transitions"},
	{Emoji: "🍱", Label: "Add or update assets"},
	{Emoji: "👔", Label: "Add or update business logic"},
	{Emoji: "🧵", Label: "Add or update code related to multithreading or concurrency"},
	{Emoji: "🦺", Label: "Add or update code related to validation"},
	{Emoji: "💡", Label: "Add or update comments in source code"},
	{Emoji: "📦️", Label: "Add or update compiled files or packages"},
	{Emoji: "🔧", Label: "Add or update configuration files"},
	{Emoji: "👥", Label: "Add or update contributor(s)"},
	{Emoji: "🔨", Label: "Add or update development scripts"},
	{Emoji: "📝", Label: "Add or update documentation"},
	{Emoji: "🩺", Label: "Add or update healthcheck"},
	{Emoji: "📄", Label: "Add or update license"},
	{Emoji: "🔊", Label: "Add or update logs"},
	{Emoji: "🔐", Label: "Add or update secrets"},
	{Emoji: "🌱", Label: "Add or update seed files"},
	{Emoji: "📸", Label: "Add or update snapshots"},
	{Emoji: "💬", Label: "Add or update text and literals"},
	{Emoji: "💄", Label: "Add or update the UI and style files"},
	{Emoji: "🏷️", Label: "Add or update types"},
	{Emoji: "💸", Label: "Add sponsorships or money related infrastructure"},
	{Emoji: "✅", Label: "Add, update, or pass tests"},
	{Emoji: "🚩", Label: "Add, update, or remove feature flags"},
	{Emoji: "🎉", Label: "Begin a project"},
	{Emoji: "🥅", Label: "Catch errors"},
	{Emoji: "🚑️", Label: "Critical hotfix"},
	{Emoji: "🧐", Label: "Data exploration/inspection"},
	{Emoji: "🚀", Label: "Deploy stuff"},
	{Emoji: "🗑️", Label: "Deprecate code that needs to be cleaned up"},
	{Emoji: "⬇️", Label: "Downgrade dependencies"},
	{Emoji: "💚", Label: "Fix CI Build"},
	{Emoji: "🐛", Label: "Fix a bug"},
	{Emoji: "🚨", Label: "Fix compiler / linter warnings"},
	{Emoji: "🔒️", Label: "Fix security or privacy issues"},
	{Emoji: "✏️", Label: "Fix typos"},
	{Emoji: "🔍️", Label: "Improve SEO"},
	{Emoji: "♿️", Label: "Improve accessibility"},
	{Emoji: "🧑‍💻", Label: "Improve developer experience"},
	{Emoji: "⚡️", Label: "Improve performance"},
	{Emoji: "🎨", Label: "Improve structure / format of the code"},
	{Emoji: "🚸", Label: "Improve user experience / usability"},
	{Emoji: "🧱", Label: "Infrastructure related changes"},
	{Emoji: "🌐", Label: "Internationalization and localization"},
	{Emoji: "💥", Label: "Introduce breaking changes"},
	{Emoji: "✨", Label: "Introduce new features"},
	{Emoji: "🏗️", Label: "Make architectural changes"},
	{Emoji: "🔀", Label: "Merge branches"},
	{Emoji: "🤡", Label: "Mock things"},
	{Emoji: "🚚", Label: "Move or rename resources (e.g.: files, paths, routes)"},
	{Emoji: "🗃️", Label: "Perform database related changes"},
	{Emoji: "⚗️", Label: "Perform experiments"},
	{Emoji: "📌", Label: "Pin dependencies to specific versions"},
	{Emoji: "♻️", Label: "Refactor code"},
	{Emoji: "🔖", Label: "Release / Version tags"},
	{Emoji: "➖", Label: "Remove a dependency"},
	{Emoji: "🔥", Label: "Remove code or files"},
	{Emoji: "⚰️", Label: "Remove dead code"},
	{Emoji: "🔇", Label: "Remove logs"},
	{Emoji: "⏪️", Label: "Revert changes"},
	{Emoji: "🩹", Label: "Simple fix for a non-critical issue"},
	{Emoji: "👽️", Label: "Update code due to external API changes"},
	{Emoji: "⬆️", Label: "Upgrade dependencies"},
	{Emoji: "🚧", Label: "Work in progress"},
	{Emoji: "🛂", Label: "Work on code related to authorization, roles and permissions"},
	{Emoji: "📱", Label: "Work on responsive design"},
	{Emoji: "💩", Label: "Write bad code that needs to be improved"},
	{Emoji: "🍻", Label: "Write code drunkenly"},
}

func GetGitmojiActions(ShowMultiCharacterGitmojis bool) []string {
	mojis := lo.Filter(gitmojis, func(g Gitmoji, _ int) bool {
		if ShowMultiCharacterGitmojis {
			return true
		}
		it := uniseg.NewGraphemes(g.Emoji)
		it.Next()
		return len(it.Runes()) <= 1
	})
	return lo.Map(mojis, func(g Gitmoji, i int) string {
		return g.Label // + "," + strconv.Itoa(uniseg.GraphemeClusterCount(g.Emoji)) + "," + strconv.Itoa(len([]rune(g.Emoji)))
	})
}

func IsGitmoji(grapheme string) bool {
	_, exists := lo.Find(gitmojis, func(g Gitmoji) bool {
		return g.Emoji == grapheme
	})
	return exists
}

func GetGitmojiByLabel(label string) string {
	gitmoji, exists := lo.Find(gitmojis, func(g Gitmoji) bool {
		return g.Label == label
	})
	if !exists {
		return ""
	}
	return gitmoji.Emoji
}
