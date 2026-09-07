/**
 * 统一确认对话框。
 *
 * 为什么不使用后端 ShowConfirmDialog（runtime.MessageDialog）：
 *  1. Windows/Linux 上 Wails 的 MessageDialog 会忽略 Buttons，只返回
 *     Yes/No/Ok/Cancel 等标准英文文本——曾经用 `result == "是"` 判定，
 *     导致所有确认框「点确定也返回 false」，功能集体失效。
 *  2. 原生对话框会夺走窗口焦点，关闭时触发 window focus，
 *     在「窗口聚焦即检测文件变更」的场景下造成弹窗死循环。
 *  3. 原生对话框样式为系统风格，与应用的简约风不一致。
 *
 * 因此统一改用 Element Plus 模态层：按钮文案可控、不夺窗口焦点、视觉统一。
 */

import { ElMessageBox } from 'element-plus'

export interface ConfirmOptions {
  title: string
  message: string
  /** 确认按钮文案，默认「确定」 */
  confirmText?: string
  /** 取消按钮文案，默认「取消」 */
  cancelText?: string
  /** 图标类型，默认 warning */
  type?: 'warning' | 'info' | 'success' | 'error'
  /** 危险操作：确认按钮显示为红色 */
  danger?: boolean
}

/** 显示确认对话框，返回用户是否确认。取消/关闭一律返回 false。 */
export async function confirmDialog(opts: ConfirmOptions): Promise<boolean> {
  try {
    await ElMessageBox.confirm(opts.message, opts.title, {
      type: opts.type ?? 'warning',
      confirmButtonText: opts.confirmText ?? '确定',
      cancelButtonText: opts.cancelText ?? '取消',
      confirmButtonClass: opts.danger ? 'el-button--danger' : '',
      closeOnClickModal: false,
      closeOnPressEscape: true,
      showClose: false,
      customClass: 'et-confirm',
    })
    return true
  } catch {
    return false
  }
}

/**
 * 三选一确认（用于「保存 / 不保存 / 取消」这类关闭确认）。
 * 返回 'save' | 'discard' | 'cancel'。
 */
export async function confirmSaveDiscard(name: string): Promise<'save' | 'discard' | 'cancel'> {
  try {
    await ElMessageBox.confirm(`是否保存对 ${name} 的更改？`, '保存更改', {
      distinguishCancelAndClose: true,
      confirmButtonText: '保存',
      cancelButtonText: '不保存',
      closeOnClickModal: false,
      showClose: false,
      customClass: 'et-confirm',
    })
    return 'save'
  } catch (e) {
    return e === 'cancel' ? 'discard' : 'cancel'
  }
}
