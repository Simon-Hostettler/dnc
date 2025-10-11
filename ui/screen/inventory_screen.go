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
	"hostettler.dev/dnc/ui/component"
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

	copper   *component.SimpleIntComponent
	silver   *component.SimpleIntComponent
	electrum *component.SimpleIntComponent
	gold     *component.SimpleIntComponent
	platinum *component.SimpleIntComponent
	itemList *list.List
}

func NewInventoryScreen(k util.KeyMap, c *models.Character) *InventoryScreen {
	return &InventoryScreen{
		keymap:    k,
		character: c,
		copper:    component.NewSimpleIntComponent(k, "CP", &c.Wallet.Copper, true, true),
		silver:    component.NewSimpleIntComponent(k, "SP", &c.Wallet.Silver, true, true),
		electrum:  component.NewSimpleIntComponent(k, "EP", &c.Wallet.Electrum, true, true),
		gold:      component.NewSimpleIntComponent(k, "GP", &c.Wallet.Gold, true, true),
		platinum:  component.NewSimpleIntComponent(k, "PP", &c.Wallet.Platinum, true, true),
	}
}

func (s *InventoryScreen) Init() tea.Cmd {
	s.populateItems()

	s.focusOn(s.copper)
	s.lastFocusedElement = s.copper

	return nil
}

func (s *InventoryScreen) populateItems() {
	if s.itemList == nil {
		s.itemList = list.NewList(s.keymap,
			list.LeftAlignedListStyle).
			WithFixedWidth(itemColWidth).
			WithViewport(itemColHeight - 2)
	}
	s.itemList.WithRows(s.GetItemRows())
}

func (s *InventoryScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case command.AppendElementMsg:
		if strings.Contains(msg.Tag, "item") {
			item_id := s.character.AddEmptyItem()
			s.populateItems()
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
	case s.copper:
		switch d {
		case command.LeftDirection:
			cmd = command.ReturnFocusToParentCmd
		case command.RightDirection:
			s.focusOn(s.silver)
		case command.DownDirection:
			s.focusOn(s.itemList)
		default:
			s.focusOn(s.copper)
		}
	case s.silver:
		switch d {
		case command.LeftDirection:
			s.focusOn(s.copper)
		case command.RightDirection:
			s.focusOn(s.electrum)
		case command.DownDirection:
			s.focusOn(s.itemList)
		default:
			s.focusOn(s.silver)
		}
	case s.electrum:
		switch d {
		case command.LeftDirection:
			s.focusOn(s.silver)
		case command.RightDirection:
			s.focusOn(s.gold)
		case command.DownDirection:
			s.focusOn(s.itemList)
		default:
			s.focusOn(s.electrum)
		}
	case s.gold:
		switch d {
		case command.LeftDirection:
			s.focusOn(s.electrum)
		case command.RightDirection:
			s.focusOn(s.platinum)
		case command.DownDirection:
			s.focusOn(s.itemList)
		default:
			s.focusOn(s.gold)
		}
	case s.platinum:
		switch d {
		case command.LeftDirection:
			s.focusOn(s.gold)
		case command.DownDirection:
			s.focusOn(s.itemList)
		default:
			s.focusOn(s.platinum)
		}
	case s.itemList:
		switch d {
		case command.LeftDirection:
			cmd = command.ReturnFocusToParentCmd
		case command.UpDirection:
			s.focusOn(s.electrum)
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
		s.populateItems()
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
	separator := util.GrayTextStyle.Width(6).Render(util.MakeVerticalSeparator(1))
	return util.DefaultBorderStyle.
		Width(util.ScreenWidth).
		AlignHorizontal(lipgloss.Left).
		Render(lipgloss.JoinHorizontal(
			lipgloss.Center,
			util.ForceWidth(s.copper.View(), 15),
			separator,
			util.ForceWidth(s.silver.View(), 15),
			separator,
			util.ForceWidth(s.electrum.View(), 15),
			separator,
			util.ForceWidth(s.gold.View(), 15),
			separator,
			util.ForceWidth(s.platinum.View(), 15),
		))
}

func RenderItemInfoRow(i *models.Item) string {
	values := []string{DrawItemPrefix(i), i.Name, DrawAttunementSlots(i.AttunementSlots)}
	values = util.Filter(values, func(s string) bool { return s != "" })
	return strings.Join(values, " ∙ ")
}

func DrawItemPrefix(i *models.Item) string {
	s := ""
	switch {
	case i.Quantity == 0 && i.Equipped:
		s = "■"
	case i.Quantity == 0 && !i.Equipped:
		s = "□"
	default:
		s = strconv.Itoa(i.Quantity)

	}
	return s
}

var AttunementSymbols []editor.EnumMapping = []editor.EnumMapping{
	{Value: 0, Label: "□□□"},
	{Value: 1, Label: "■□□"},
	{Value: 2, Label: "■■□"},
	{Value: 3, Label: "■■■"},
}

func DrawAttunementSlots(used int) string {
	s := strings.Repeat("■", used)
	s += strings.Repeat("□", 3-used)
	return s
}
