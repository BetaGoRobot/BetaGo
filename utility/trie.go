package utility

import (
	"regexp"
	"strings"
)

// ReplaceMentionsRawText 在文本中查找所有 @xxx 格式的用户名，并根据提供的 map 进行替换。
// @ 前面必须是文本开头或空格。
// xxx 必须是一个完整的单词（后面是空格、标点符号或文本结尾）。
func ReplaceMentionsRawText(text string, replacements map[string]string) string {
	// 正则表达式: (?:^|\s)@(\w+)\b
	// (?:^|\s)  - 非捕获组，匹配字符串开头或空格
	// @         - 匹配'@'符号
	// (\w+)     - 捕获组，匹配一个或多个单词字符（这是 key）
	// \b        - 单词边界，确保我们匹配的是一个完整的 "单词"
	re := regexp.MustCompile(`@([\p{Han}\w]+)`)

	// FindAllStringSubmatchIndex 会返回所有匹配项的索引位置
	// 每个匹配项是一个数组：[完整匹配开始, 完整匹配结束, 第一个捕获组开始, 第一个捕获组结束]
	matches := re.FindAllStringSubmatchIndex(text, -1)

	// 如果没有找到匹配项，直接返回原文
	if len(matches) == 0 {
		return text
	}

	var builder strings.Builder
	lastIndex := 0

	for _, match := range matches {
		// match[2] 和 match[3] 是捕获组 (\w+) 的开始和结束索引，也就是 'xxx'
		key := text[match[2]:match[3]]

		// 从 map 中查找替换值
		value, found := replacements[key]

		// 找到要替换的整个模式 "@xxx" 的起始位置
		// 注意: 捕获组 'xxx' 的起始位置 (match[2]) 的前一个字符就是 '@'
		mentionStartIndex := match[2] - 1

		// 1. 先将上一个匹配项到当前匹配项之间的文本追加进来
		builder.WriteString(text[lastIndex:mentionStartIndex])

		if found {
			// 2. 如果在 map 中找到了 key，则追加替换后的 value
			builder.WriteString(value)
		} else {
			// 3. 如果没找到，则将原始的 "@xxx" 追加回来
			builder.WriteString(text[mentionStartIndex:match[3]])
		}

		// 4. 更新 lastIndex，为下一次循环做准备
		lastIndex = match[3]
	}

	// 追加最后一个匹配项到字符串末尾的剩余文本
	builder.WriteString(text[lastIndex:])

	return builder.String()
}

// TrieNode 定义了字典树的节点
type TrieNode struct {
	children    map[rune]*TrieNode
	isEndOfWord bool
	replacement string // 在单词结束节点存储替换文本
}

// NewTrieNode 创建一个新的字典树节点
func NewTrieNode() *TrieNode {
	return &TrieNode{
		children:    make(map[rune]*TrieNode),
		isEndOfWord: false,
		replacement: "",
	}
}

// BuildTrie 从一个 map[string]string 构建字典树。
// map 的 key 是要查找的关键词，value 是对应的替换文本。
func BuildTrie(wordList map[string]string) *TrieNode {
	root := NewTrieNode()
	for key, value := range wordList {
		node := root
		// 将关键词（rune切片）插入字典树
		for _, r := range []rune(key) {
			if _, found := node.children[r]; !found {
				node.children[r] = NewTrieNode()
			}
			node = node.children[r]
		}
		// 标记单词结束，并存储替换值
		node.isEndOfWord = true
		node.replacement = value
	}
	return root
}

// ReplaceMentionsWithTrie 使用预构建的字典树来查找并替换文本中的 @mentions。
func (root *TrieNode) ReplaceMentionsWithTrie(text string) string {
	if root == nil || len(root.children) == 0 {
		return text
	}

	var builder strings.Builder
	runes := []rune(text) // 使用 rune 以正确处理多字节字符（如中文）
	lastIndex := 0

	for i := 0; i < len(runes); i++ {
		// 寻找 @ 符号
		if runes[i] != '@' {
			continue
		}

		// 从 @ 符号的下一个字符开始在字典树中搜索
		node := root
		matchEndIndex := -1
		var foundReplacement string

		for j := i + 1; j < len(runes); j++ {
			char := runes[j]
			child, found := node.children[char]
			if !found {
				// 如果当前字符在字典树中没有对应的子节点，则停止搜索
				break
			}

			node = child
			// 如果当前节点是一个单词的结尾，记录下这个可能的匹配
			// 我们继续搜索，以找到最长的匹配项
			if node.isEndOfWord {
				matchEndIndex = j
				foundReplacement = node.replacement
			}
		}

		// 如果找到了一个或多个匹配项，matchEndIndex 会记录最长匹配的结束位置
		if matchEndIndex != -1 {
			// 1. 将上一个匹配结束位置到当前 @ 符号前的文本追加进来
			builder.WriteString(string(runes[lastIndex:i]))
			// 2. 追加替换后的文本
			builder.WriteString(foundReplacement)
			// 3. 更新索引，跳过已处理的 @mention
			i = matchEndIndex
			lastIndex = i + 1
		}
	}

	// 4. 追加最后一个匹配项到字符串末尾的剩余文本
	if lastIndex < len(runes) {
		builder.WriteString(string(runes[lastIndex:]))
	}

	return builder.String()
}
