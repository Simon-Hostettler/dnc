package screen

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/repository"
	"hostettler.dev/dnc/ui/component"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/list"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

var (
	itemColHeight = 30
	itemColWidth  = styles.ScreenWidth - 8
)

type InventoryScreen struct {
	keymap    util.KeyMap
	character *repository.CharacterAggregate

	lastFocusedElement FocusableModel
	focusedElement     FocusableModel

	copper   *component.SimpleIntComponent
	silver   *component.SimpleIntComponent
	electrum *component.SimpleIntComponent
	gold     *component.SimpleIntComponent
	platinum *component.SimpleIntComponent
	itemList *list.List
}

func NewInventoryScreen(k util.KeyMap, c *repository.CharacterAggregate) *InventoryScreen {
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
	s.CreateItemRows()
}

func (s *InventoryScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case command.AppendElementMsg:
		if strings.Contains(msg.Tag, "item") {
			item_id := s.character.AddEmptyItem()
			s.populateItems()
			cmd = editor.SwitchToEditorCmd(
				s.getItemRow(item_id).Editors(),
			)
		}
	case command.FocusNextElementMsg:
		s.moveFocus(msg.Direction)
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

	content := styles.DefaultBorderStyle.
		Width(styles.ScreenWidth).
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

func (s *InventoryScreen) CreateItemRows() {
	rows := []list.Row{}
	for i := range s.character.Items {
		item := &s.character.Items[i]
		rows = append(rows, list.NewStructRow(s.keymap, item,
			renderItemInfoRow,
			createItemEditors(s.keymap, item),
		).WithDestructor(deleteItemCallback(s, item)).
			WithReader(renderFullItemInfo))
	}
	rows = append(rows, list.NewAppenderRow(s.keymap, "item"))
	s.itemList.WithRows(rows)
}

func (s *InventoryScreen) getItemRow(id uuid.UUID) list.Row {
	for _, r := range s.itemList.Content() {
		switch r := r.(type) {
		case *list.StructRow[models.ItemTO]:
			if r.Value().ID == id {
				return r
			}
		}
	}
	return nil
}

func deleteItemCallback(s *InventoryScreen, i *models.ItemTO) func() tea.Cmd {
	return func() tea.Cmd {
		s.character.DeleteItem(i.ID)
		s.populateItems()
		return command.WriteBackRequest
	}
}

func createItemEditors(k util.KeyMap, item *models.ItemTO) []editor.ValueEditor {
	return []editor.ValueEditor{
		editor.NewStringEditor(k, "Name", &item.Name),
		editor.NewEnumEditor(k, models.EquippedSymbols, "Equipped", &item.Equipped),
		editor.NewEnumEditor(k, models.AttunementSymbols, "Attunement Slots", &item.AttunementSlots),
		editor.NewIntEditor(k, "Quantity", &item.Quantity),
		editor.NewTextEditor(k, "Description", &item.Description),
	}
}

func (s *InventoryScreen) RenderInventoryScreenTopBar() string {
	separator := styles.GrayTextStyle.Width(6).Render(styles.MakeVerticalSeparator(1))
	return styles.DefaultBorderStyle.
		Width(styles.ScreenWidth).
		AlignHorizontal(lipgloss.Left).
		Render(lipgloss.JoinHorizontal(
			lipgloss.Center,
			styles.ForceWidth(s.copper.View(), 15),
			separator,
			styles.ForceWidth(s.silver.View(), 15),
			separator,
			styles.ForceWidth(s.electrum.View(), 15),
			separator,
			styles.ForceWidth(s.gold.View(), 15),
			separator,
			styles.ForceWidth(s.platinum.View(), 15),
		))
}

func renderItemInfoRow(i *models.ItemTO) string {
	values := []string{drawItemPrefix(i), i.Name, styles.PrettyAttunementSlots(i.AttunementSlots)}
	values = util.Filter(values, func(s string) bool { return s != "" })
	return strings.Join(values, " ∙ ")
}

func renderFullItemInfo(i *models.ItemTO) string {
	separator := styles.MakeHorizontalSeparator(styles.SmallScreenWidth-4, 1)
	content := strings.Join(
		[]string{
			i.Name,
			separator,
			"Equipped: " + drawItemPrefix(i),
			separator,
			"Attunement slots required: " + styles.PrettyAttunementSlots(i.AttunementSlots),
			separator,
			"Quantity: " + strconv.Itoa(i.Quantity),
			separator,
			i.Description,
		},
		"\n")
	return styles.DefaultTextStyle.
		AlignHorizontal(lipgloss.Left).
		Render(content)
}

func drawItemPrefix(i *models.ItemTO) string {
	s := ""
	switch i.Equipped {
	case models.Equipped:
		s = "■"
	case models.NotEquipped:
		s = "□"
	default:
		s = strconv.Itoa(i.Quantity)

	}
	return s
}
