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
	FocusManager

	copper   *component.SimpleComponent[int]
	silver   *component.SimpleComponent[int]
	electrum *component.SimpleComponent[int]
	gold     *component.SimpleComponent[int]
	platinum *component.SimpleComponent[int]
	itemList *list.List

	itemRows *Collection[models.ItemTO]
}

func NewInventoryScreen(k util.KeyMap, c *repository.CharacterAggregate) *InventoryScreen {
	s := &InventoryScreen{
		keymap:    k,
		character: c,
		copper:    component.NewSimpleIntComponent(k, "CP", &c.Wallet.Copper, true, true),
		silver:    component.NewSimpleIntComponent(k, "SP", &c.Wallet.Silver, true, true),
		electrum:  component.NewSimpleIntComponent(k, "EP", &c.Wallet.Electrum, true, true),
		gold:      component.NewSimpleIntComponent(k, "GP", &c.Wallet.Gold, true, true),
		platinum:  component.NewSimpleIntComponent(k, "PP", &c.Wallet.Platinum, true, true),
		itemList: list.NewList(k, list.LeftAlignedListStyle).
			WithFixedWidth(itemColWidth).
			WithViewport(itemColHeight - 2).
			WithSearch(),
	}
	s.itemRows = NewCollection(k, s.itemList,
		func() []*models.ItemTO { return util.Pointers(s.character.Items) },
		func(i *models.ItemTO) uuid.UUID { return i.ID },
		s.character.AddEmptyItem,
		s.character.DeleteItem,
		func(item *models.ItemTO) *list.StructRow[models.ItemTO] {
			return list.NewStructRow(s.keymap, item, renderItemInfoRow,
				createItemEditors(s.keymap, item)).
				WithReader(renderFullItemInfo).
				WithSearchText(itemSearchText)
		},
	)
	return s
}

func (s *InventoryScreen) Init() tea.Cmd {
	s.itemRows.Repopulate()

	s.wireFocusGraph()

	return nil
}

func (s *InventoryScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case command.FocusNextElementMsg:
		s.MoveFocus(msg.Direction)
	case tea.KeyPressMsg:
		cmd = RouteKey(s.focusedElement, msg, s.keymap, s.MoveFocus)
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

func (s *InventoryScreen) wireFocusGraph() {
	s.Wire(FocusGraph{
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
	}, s.copper)
}

func createItemEditors(k util.KeyMap, item *models.ItemTO) []editor.ValueEditor {
	return []editor.ValueEditor{
		editor.NewStringEditor(k, "Name", &item.Name),
		editor.NewEnumEditor(k, styles.IsEquippableSymbols, "Equippable", &item.IsEquippable),
		editor.NewEnumEditor(k, styles.EquippedSymbols, "Equipped", &item.Equipped).
			WithDisabledWhen(func() bool { return item.IsEquippable == 0 }),
		editor.NewEnumEditor(k, styles.AttunementSymbols, "Attunement Slots", &item.AttunementSlots),
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

func itemSearchText(i *models.ItemTO) string {
	return i.Name + " " + i.Description
}

func renderItemInfoRow(i *models.ItemTO) string {
	values := []string{drawItemPrefix(i), i.Name, styles.PrettyAttunementSlots(i.AttunementSlots)}
	values = util.Filter(values, func(s string) bool { return s != "" })
	return strings.Join(values, " ∙ ")
}

func renderFullItemInfo(i *models.ItemTO) string {
	separator := styles.MakeHorizontalSeparator(styles.SmallScreenWidth-4, 1)
	equippedValue := "Not Equippable"
	if i.IsEquippable == 1 {
		equippedValue = drawItemPrefix(i)
	}
	content := strings.Join(
		[]string{
			i.Name,
			separator,
			"Equipped: " + equippedValue,
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
	if i.IsEquippable == 0 {
		return strconv.Itoa(i.Quantity)
	}
	if i.Equipped == 1 {
		return "■"
	}
	return "□"
}
