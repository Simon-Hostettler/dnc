package screen

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	itemColWidth  = styles.ScreenWidth - 10
)

type InventoryScreen struct {
	keymap    util.KeyMap
	character *repository.CharacterAggregate

	focusGraph         FocusGraph
	lastFocusedElement FocusableModel
	focusedElement     FocusableModel

	copper   *component.SimpleComponent[int]
	silver   *component.SimpleComponent[int]
	electrum *component.SimpleComponent[int]
	gold     *component.SimpleComponent[int]
	platinum *component.SimpleComponent[int]
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
	s.wireFocusGraph()

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
	case tea.KeyPressMsg:
		cmd = RouteKey(s.focusedElement, msg, s.keymap, s.moveFocus)
	}
	return s, cmd
}

func (s *InventoryScreen) View() tea.View {
	topbar := s.RenderInventoryScreenTopBar()
	renderedItems := s.itemList.View().Content

	content := styles.DefaultBorderStyle.
		Width(styles.ScreenWidth).
		Height(itemColHeight).
		Render(renderedItems)
	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, topbar, content))
}

func (s *InventoryScreen) focusOn(m FocusableModel) {
	s.focusedElement = m
	m.Focus()
}

func (s *InventoryScreen) wireFocusGraph() {
	s.focusGraph = FocusGraph{
		s.copper: {
			command.LeftDirection:  Emit(command.ReturnFocusToParentCmd),
			command.RightDirection: To(s.silver),
			command.DownDirection:  To(s.itemList),
		},
		s.silver: {
			command.LeftDirection:  To(s.copper),
			command.RightDirection: To(s.electrum),
			command.DownDirection:  To(s.itemList),
		},
		s.electrum: {
			command.LeftDirection:  To(s.silver),
			command.RightDirection: To(s.gold),
			command.DownDirection:  To(s.itemList),
		},
		s.gold: {
			command.LeftDirection:  To(s.electrum),
			command.RightDirection: To(s.platinum),
			command.DownDirection:  To(s.itemList),
		},
		s.platinum: {
			command.LeftDirection: To(s.gold),
			command.DownDirection: To(s.itemList),
		},
		s.itemList: {
			command.LeftDirection: Emit(command.ReturnFocusToParentCmd),
			command.UpDirection:   To(s.electrum),
		},
	}
}

func (s *InventoryScreen) moveFocus(d command.Direction) tea.Cmd {
	edge, ok := s.focusGraph[s.focusedElement][d]
	if !ok {
		return nil
	}
	target, cmd := edge()
	if target != nil {
		s.Blur()
		s.focusOn(target)
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
	return list.FindStructRow(s.itemList.Content(), func(i *models.ItemTO) bool { return i.ID == id })
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
			styles.ForceWidth(s.copper.View().Content, 15),
			separator,
			styles.ForceWidth(s.silver.View().Content, 15),
			separator,
			styles.ForceWidth(s.electrum.View().Content, 15),
			separator,
			styles.ForceWidth(s.gold.View().Content, 15),
			separator,
			styles.ForceWidth(s.platinum.View().Content, 15),
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
