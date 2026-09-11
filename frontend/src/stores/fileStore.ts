import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { TreeNode, FileInfo } from '@/types'

export const useFileStore = defineStore('file', () => {
  // State
  const currentDirectory = ref<string | null>(null)
  const fileTree = ref<TreeNode | null>(null)
  const expandedPaths = ref<Set<string>>(new Set())
  const selectedPath = ref<string | null>(null)
  const isLoading = ref(false)

  // Computed
  const hasDirectory = computed(() => currentDirectory.value !== null)

  // Actions
  function setDirectory(path: string | null) {
    currentDirectory.value = path
  }

  /** 收集树中所有节点路径（用于裁剪失效的展开/选中状态） */
  function collectPaths(node: TreeNode | null, out: Set<string>) {
    if (!node) return
    out.add(node.path)
    for (const c of node.children || []) collectPaths(c, out)
  }

  function setFileTree(tree: TreeNode | null) {
    fileTree.value = tree
    if (!tree) {
      expandedPaths.value = new Set()
      selectedPath.value = null
      return
    }
    // 刷新/删除/重命名后，旧路径会残留在 expandedPaths 里，
    // 造成「展开了不存在的目录」「选中项无高亮」等不一致，这里按新树裁剪。
    const valid = new Set<string>()
    collectPaths(tree, valid)
    const next = new Set<string>()
    for (const p of expandedPaths.value) {
      if (valid.has(p)) next.add(p)
    }
    next.add(tree.path) // 根节点默认展开
    expandedPaths.value = next
    if (selectedPath.value && !valid.has(selectedPath.value)) {
      selectedPath.value = null
    }
  }

  function expandPath(path: string) {
    expandedPaths.value.add(path)
  }

  function collapsePath(path: string) {
    expandedPaths.value.delete(path)
  }

  function togglePath(path: string) {
    if (expandedPaths.value.has(path)) {
      expandedPaths.value.delete(path)
    } else {
      expandedPaths.value.add(path)
    }
  }

  function isExpanded(path: string): boolean {
    return expandedPaths.value.has(path)
  }

  function setSelectedPath(path: string | null) {
    selectedPath.value = path
  }

  function setLoading(loading: boolean) {
    isLoading.value = loading
  }

  // Find node by path
  function findNode(path: string, node?: TreeNode): TreeNode | null {
    const searchNode = node || fileTree.value
    if (!searchNode) return null

    if (searchNode.path === path) {
      return searchNode
    }

    if (searchNode.children) {
      for (const child of searchNode.children) {
        const found = findNode(path, child)
        if (found) return found
      }
    }

    return null
  }

  // Expand path to reveal a file
  function expandToPath(path: string) {
    if (!fileTree.value) return

    const parts = path.split(/[/\\]/)
    let currentPath = ''

    for (let i = 0; i < parts.length - 1; i++) {
      currentPath = currentPath ? `${currentPath}/${parts[i]}` : parts[i]
      expandedPaths.value.add(currentPath)
    }
  }

  return {
    // State
    currentDirectory,
    fileTree,
    expandedPaths,
    selectedPath,
    isLoading,
    // Computed
    hasDirectory,
    // Actions
    setDirectory,
    setFileTree,
    expandPath,
    collapsePath,
    togglePath,
    isExpanded,
    setSelectedPath,
    setLoading,
    findNode,
    expandToPath,
  }
})