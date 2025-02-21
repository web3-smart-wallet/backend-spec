package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// 添加页面状态常量
const (
	passwordPage = iota // 添加密码页面作为第一个状态
	menuPage
	deployPage
	deployContractPage
	airdropPage
	upLoadPage
	checkTotalPage
)

// 其他常见常量定义
const (
	urlPattern     = `^(http|https)://[a-zA-Z0-9\-._~:/?#\[\]@!$&'()*+,;=]+$` // 改进URL验证模式
	password       = "123456"
	maxNFTIDLength = 10  // NFT ID最大长度
	maxURLLength   = 255 // URL最大长度
)

// 在文件开头添加自定义消息类型
type airdropMsg struct {
	nftID  string
	nftURL string
}

// 添加错误消息类型
type errorMsg struct {
	err error
}

// 添加成功消息类型
type successMsg struct {
	message string
}

type model struct {
	choices        []string // 菜单选项
	cursor         int      // 当前光标位置
	selected       int      // 当前选中的选项
	currentPage    int      // 当前页面状态
	deployChoices  []string // 部署合约选项
	deployCursor   int      // 部署合约光标位置
	nftInput       string   // 输入框内容
	graphURL       string   // Graph URL输入内容
	inputMode      string   // 输入模式：'nft' 或 'url'
	inputCursor    int      // 输入框光标位置
	password       string   // 用户输入的密码
	authenticated  bool     // 验证状态
	errorMessage   string   // 错误消息
	successMessage string   // 成功消息
	loading        bool     // 加载状态
}

func initialModel() model {
	return model{
		choices:        []string{"Deploy Contract", "AirDrop NFT", "Check Total NFT", "Settings"},
		cursor:         0,
		selected:       0,
		currentPage:    passwordPage,
		deployChoices:  []string{"Mint New NFT", "Deploy ERC1155"},
		deployCursor:   0,
		nftInput:       "",
		graphURL:       "",
		inputMode:      "nft",
		inputCursor:    0,
		password:       "",
		authenticated:  false,
		errorMessage:   "",
		successMessage: "",
		loading:        false,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case errorMsg:
		m.errorMessage = msg.err.Error()
		return m, nil

	case successMsg:
		m.successMessage = msg.message
		return m, nil

	case tea.KeyMsg:
		// 清除错误和成功消息
		m.errorMessage = ""
		m.successMessage = ""

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc": // 添加返回功能
			if m.currentPage != menuPage {
				if m.currentPage == airdropPage && m.inputMode == "url" {
					m.graphURL = ""
					m.inputMode = "nft"
				} else if m.currentPage == upLoadPage {
					m.currentPage = airdropPage
				} else {
					m.currentPage = menuPage
					m.nftInput = ""
				}

			}
		case "up":
			if m.currentPage == menuPage && m.cursor > 0 {
				m.cursor--
			}
			if m.currentPage == deployPage && m.deployCursor > 0 {
				m.deployCursor--
			}
		case "down":
			if m.currentPage == menuPage && m.cursor < len(m.choices)-1 {
				m.cursor++
			}
			if m.currentPage == deployPage && m.deployCursor < len(m.deployChoices)-1 {
				m.deployCursor++
			}
		case "backspace":
			if m.currentPage == airdropPage {
				if m.inputMode == "nft" && len(m.nftInput) > 0 {
					m.nftInput = m.nftInput[:len(m.nftInput)-1]

				} else if m.inputMode == "url" && len(m.graphURL) > 0 {
					m.graphURL = m.graphURL[:len(m.graphURL)-1]
				}
			} else if m.currentPage == passwordPage && len(m.password) > 0 {
				m.password = m.password[:len(m.password)-1]
			}
		case "enter":
			if m.currentPage == passwordPage {
				if password == m.password {
					m.authenticated = true
					m.currentPage = menuPage
					m.password = ""
					return m, func() tea.Msg {
						return successMsg{message: "登录成功！"}
					}
				}
				return m, func() tea.Msg {
					return errorMsg{err: fmt.Errorf("密码错误")}
				}
			} else if m.currentPage == menuPage {
				// 根据选择切换到对应页面
				switch m.cursor {
				case 0:
					m.currentPage = deployPage
				case 1:
					m.currentPage = airdropPage
					m.nftInput = ""
				case 2:
					m.currentPage = checkTotalPage
				case 3:
					m.currentPage = passwordPage
				}
			} else if m.currentPage == airdropPage {
				if m.inputMode == "nft" {
					if len(m.nftInput) == 0 {
						return m, func() tea.Msg {
							return errorMsg{err: fmt.Errorf("NFT ID 不能为空")}
						}
					}
					if len(m.nftInput) > maxNFTIDLength {
						return m, func() tea.Msg {
							return errorMsg{err: fmt.Errorf("NFT ID 太长")}
						}
					}
					m.inputMode = "url"
					return m, nil
				}

				if m.inputMode == "url" {
					if len(m.graphURL) == 0 {
						return m, func() tea.Msg {
							return errorMsg{err: fmt.Errorf("URL 不能为空")}
						}
					}
					if len(m.graphURL) > maxURLLength {
						return m, func() tea.Msg {
							return errorMsg{err: fmt.Errorf("URL 太长")}
						}
					}
					matched, _ := regexp.MatchString(urlPattern, m.graphURL)
					if !matched {
						return m, func() tea.Msg {
							return errorMsg{err: fmt.Errorf("无效的 URL 格式")}
						}
					}
					return m, func() tea.Msg {
						return airdropMsg{nftID: m.nftInput, nftURL: m.graphURL}
					}
				}
			} else if m.currentPage == upLoadPage {
				// 上传文件
				m.currentPage = menuPage
				return m, nil

			} else if m.currentPage == deployPage {
				switch m.deployCursor {
				case 0:
					m.currentPage = deployContractPage
					// case 1:
					// 	m.currentPage = checkTotalPage
				}
			}

		default:
			if m.currentPage == airdropPage {
				if m.inputMode == "nft" {
					// 只接受数字输入
					if _, err := strconv.Atoi(msg.String()); err == nil {
						m.nftInput += msg.String()
					}
				} else if m.inputMode == "url" {
					// 使用正则表达式验证URL字符
					if matched, _ := regexp.MatchString(urlPattern, msg.String()); matched {
						m.graphURL += msg.String()
					}
				}
			} else if m.currentPage == passwordPage {
				m.password += msg.String()
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	// 显示错误消息
	if m.errorMessage != "" {
		s.WriteString(fmt.Sprintf("\n❌ 错误: %s\n\n", m.errorMessage))
	}

	// 显示成功消息
	if m.successMessage != "" {
		s.WriteString(fmt.Sprintf("\n✅ %s\n\n", m.successMessage))
	}

	// 显示加载状态
	if m.loading {
		s.WriteString("正在处理...\n\n")
	}

	switch m.currentPage {
	case passwordPage:
		return m.passwordView()
	case menuPage:
		return m.menuView()
	case deployPage:
		return m.deployView()
	case deployContractPage:
		return m.deployContractView()
	case airdropPage:
		return m.airdropView()
	case upLoadPage:
		return m.upLoadView()
	case checkTotalPage:
		return m.checkTotalView()
	default:
		return "未知页面\n"
	}
}

// 将原来的 View 方法改名为 menuView
func (m model) menuView() string {
	s := "请选择操作:\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}
	s += "\n主菜单.\n"
	s += "按 Ctrl+C 退出程序.\n"
	return s
}

// 添加新的页面视图
func (m model) deployView() string {
	s := "部署合约页面\n------------\n"

	for i, choice := range m.deployChoices {
		cursor := " "
		if m.deployCursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\n按 ESC 返回主菜单.\n"
	s += "按 Ctrl+C 退出程序.\n"
	return s
}

func (m model) deployContractView() string {
	return `
Mint NFT 页面
--------------
这里是 Mint NFT 的界面

按 ESC 返回主菜单
按 Ctrl+C 退出程序
`
}

// 修改 airdropView 方法，添加消息显示
func (m model) airdropView() string {
	s := "空投 NFT 页面\n"
	s += "--------------\n\n"

	if m.inputMode == "nft" {
		s += "请输入要空投的 NFT 编号：\n"
		s += fmt.Sprintf("> %s", m.nftInput)
		if len(m.nftInput) == 0 {
			s += "_"
		}
		s += "\n\n按 Enter 继续"
		s += "\n按 ESC 返回主菜单\n"
		s += "按 Ctrl+C 退出程序\n"
	} else {
		s += fmt.Sprintf("NFT 编号: %s\n\n", m.nftInput)
		s += "请输入 Graph URL：\n"
		s += fmt.Sprintf("> %s", m.graphURL)
		if len(m.graphURL) == 0 {
			s += "_"
		}
		s += "\n\n按 Enter 确认空投"
		s += "\n按 ESC 重新输入 NFT 编号\n"
		s += "按 Ctrl+C 退出程序\n"
	}

	return s
}

func (m model) upLoadView() string {
	return `
上传文件页面
--------------
这里是上传文件的界面

按 ESC 返回上一页
按 Ctrl+C 退出程序
`
}

func (m model) checkTotalView() string {
	return `
查看 NFT 总量页面
----------------
这里是查看 NFT 总量的界面

按 ESC 返回主菜单
按 Ctrl+C 退出程序
`
}

// 添加密码验证视图
func (m model) passwordView() string {
	s := "请输入密码:\n\n"
	s += "> " + strings.Repeat("*", len(m.password))
	if len(m.password) == 0 {
		s += "_"
	}
	s += "\n\n按 Enter 确认\n"
	s += "按 Ctrl+C 退出程序\n"
	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	_, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
