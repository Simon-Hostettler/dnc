package screen

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/ui/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/list"
	"hostettler.dev/dnc/ui/util"
)

var (
	itemColHeight = 30
	itemColWidth  = util.ScreenWidth - 8
)

type InventoryScreen struct {
	keymap    util.KeyMap
	character *models.Character

	lastFocusedElement FocusableModel
	focusedElement     FocusableModel

	wallet   *list.List
	itemList *list.List
}

func NewInventoryScreen(k util.KeyMap, c *models.Character) *InventoryScreen {
	return &InventoryScreen{
		keymap:    k,
		character: c,
	}
}

func (s *InventoryScreen) Init() tea.Cmd {
	s.focusOn(s.wallet)
	s.lastFocusedElement = s.wallet

	s.populateLists()

	return nil
}

func (s *InventoryScreen) populateLists() {
	if s.itemList == nil {
		s.itemList = list.NewList(s.keymap,
			list.LeftAlignedListStyle).
			WithFixedWidth(itemColWidth).
			WithViewport(itemColHeight - 2)
	}
	if s.wallet == nil {
		s.wallet = list.NewList(s.keymap, list.LeftAlignedListStyle).
			WithFixedWidth(itemColWidth)
	}
	s.itemList.WithRows(s.GetItemRows())
	s.wallet.WithRows(s.GetWalletRows())
}

func (s *InventoryScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case command.AppendElementMsg:
		if strings.Contains(msg.Tag, "item") {
			item_id := s.character.AddEmptyItem()
			s.populateLists()
			cmd = editor.SwitchToEditorCmd(
				s.character,
				s.getItemRow(item_id).Editors(),
			)
		}
	case command.FocusNextElementMsg:
		s.moveFocus(msg.Direction)
	case editor.EditValueMsg:
		cmd = editor.SwitchToEditorCmd(s.character, msg.Editors)
	case tea.KeyMsg:
		switch s.focusedElement.(type) {
		case *list.List:
			switch {
			case key.Matches(msg, s.keymap.Right):
				cmd = s.moveFocus(command.RightDirection)
			case key.Matches(msg, s.keymap.Left):
				cmd = s.moveFocus(command.LeftDirection)
			default:
				_, cmd = s.focusedElement.Update(msg)
			}
		default:
			switch {
			case key.Matches(msg, s.keymap.Right):
				cmd = s.moveFocus(command.RightDirection)
			case key.Matches(msg, s.keymap.Left):
				cmd = s.moveFocus(command.LeftDirection)
			case key.Matches(msg, s.keymap.Up):
				cmd = s.moveFocus(command.UpDirection)
			case key.Matches(msg, s.keymap.Down):
				cmd = s.moveFocus(command.DownDirection)
			default:
				_, cmd = s.focusedElement.Update(msg)
			}
		}
	}
	return s, cmd
}

func (s *InventoryScreen) View() string {
	topbar := s.RenderInventoryScreenTopBar()
	renderedItems := s.itemList.View()

	content := util.DefaultBorderStyle.
		Width(util.ScreenWidth).
		Height(spellColHeight).
		Render(renderedItems)
	return lipgloss.JoinVertical(lipgloss.Left, topbar, content)
}

func (s *InventoryScreen) focusOn(m FocusableModel) {
	s.focusedElement = m
	m.Focus()
}

func (s *InventoryScreen) moveFocus(d command.Direction) tea.Cmd {
	var cmd tea.Cmd
	s.Blur()

	switch s.lastFocusedElement {
	case s.wallet:
		switch d {
		case command.LeftDirection:
			cmd = command.ReturnFocusToParentCmd
		case command.DownDirection:
			s.focusOn(s.itemList)
		default:
			s.focusOn(s.wallet)
		}
	case s.itemList:
		switch d {
		case command.LeftDirection:
			cmd = command.ReturnFocusToParentCmd
		case command.UpDirection:
			s.focusOn(s.wallet)
		default:
			s.focusOn(s.itemList)
		}
	}
	return cmd
}

func (s *InventoryScreen) Focus() {
	s.focusOn(s.lastFocusedElement)
}

func (s *InventoryScreen) Blur() {
	if s.focusedElement != nil {
		s.focusedElement.Blur()
		s.lastFocusedElement = s.focusedElement
	}

	s.focusedElement = nil
}

func (s *InventoryScreen) GetWalletRows() []list.Row {
	rows := []list.Row{
		list.NewLabeledIntRow(s.keymap, "CP", &s.character.Wallet.Copper,
			editor.NewIntEditor(s.keymap, "CP", &s.character.Wallet.Copper)),
		list.NewLabeledIntRow(s.keymap, "SP", &s.character.Wallet.Silver,
			editor.NewIntEditor(s.keymap, "SP", &s.character.Wallet.Silver)),
		list.NewLabeledIntRow(s.keymap, "EP", &s.character.Wallet.Electrum,
			editor.NewIntEditor(s.keymap, "EP", &s.character.Wallet.Electrum)),
		list.NewLabeledIntRow(s.keymap, "GP", &s.character.Wallet.Gold,
			editor.NewIntEditor(s.keymap, "GP", &s.character.Wallet.Gold)),
		list.NewLabeledIntRow(s.keymap, "PP", &s.character.Wallet.Platinum,
			editor.NewIntEditor(s.keymap, "PP", &s.character.Wallet.Platinum)),
	}
	return rows
}

func (s *InventoryScreen) GetItemRows() []list.Row {
	rows := []list.Row{}
	for i := range s.character.Equipment {
		item := &s.character.Equipment[i]
		rows = append(rows, list.NewStructRow(s.keymap, item,
			RenderItemInfoRow,
			CreateItemEditors(s.keymap, item),
		).WithDestructor(command.InventoryScreenIndex, DeleteItemCallback(s, item)))
	}
	rows = append(rows, list.NewAppenderRow(s.keymap, "item"))
	return rows
}

func (s *InventoryScreen) getItemRow(id uuid.UUID) list.Row {
	for _, r := range s.itemList.Content() {
		switch r := r.(type) {
		case *list.StructRow[models.Item]:
			if r.Value().Id == id {
				return r
			}
		}
	}
	return nil
}

func DeleteItemCallback(s *InventoryScreen, i *models.Item) func() tea.Cmd {
	return func() tea.Cmd {
		s.character.DeleteItem(i.Id)
		s.populateLists()
		return command.SaveToFileCmd(s.character)
	}
}

func CreateItemEditors(k util.KeyMap, item *models.Item) []editor.ValueEditor {
	return []editor.ValueEditor{
		editor.NewStringEditor(k, "Name", &item.Name),
		editor.NewBooleanEditor(k, "Equipped", &item.Equipped),
		editor.NewEnumEditor(k, AttunementSymbols, "Attunement Slots", &item.AttunementSlots),
		editor.NewIntEditor(k, "Quantity", &item.Quantity),
		editor.NewStringEditor(k, "Description", &item.Description),
	}
}

func (s *InventoryScreen) RenderInventoryScreenTopBar() string {
	return util.DefaultBorderStyle.
		Width(util.ScreenWidth).
		Render(s.wallet.View())
}

func RenderItemInfoRow(i *models.Item) string {
	values := []string{RenderItemPrefix(i), i.Name, RenderAttunementSlots(i.AttunementSlots)}
	values = util.Filter(values, func(s string) bool { return s != "" })
	return strings.Join(values, " ∙ ")
}

func RenderItemPrefix(i *models.Item) string {
	s := ""
	switch {
	case i.Quantity == 0 && i.Equipped:
		s = "■"
	case i.Quantity == 0 && !i.Equipped:
		s = "□"
	default:
		s = strconv.Itoa(i.Quantity)

	}
	return util.DefaultTextStyle.Render(s)
}

var AttunementSymbols []editor.EnumMapping = []editor.EnumMapping{
	{Value: 0, Label: "□□□"},
	{Value: 1, Label: "■□□"},
	{Value: 2, Label: "■■□"},
	{Value: 3, Label: "■■■"},
}

func RenderAttunementSlots(used int) string {
	s := strings.Repeat("■", used)
	s += strings.Repeat("□", 3-used)
	return util.DefaultTextStyle.Render(s)
}
