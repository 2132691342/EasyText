import { nextTick } from 'vue'

/**
 * 右键菜单视口夹取。
 *
 * 各组件的右键菜单此前直接以 e.clientX/clientY 作为 fixed 定位坐标，
 * 鼠标贴近窗口右缘/下缘时菜单会被裁剪出视口外，位于底部的菜单项
 * 根本点不到（FileTree 菜单十余项只露出第一项的实测场景）。
 *
 * 在菜单 visible=true 之后的 nextTick 调用：此时菜单已按内容渲染出
 * 真实尺寸，把坐标夹取到视口内即可。
 *
 * @param el   菜单根元素（v-if + ref 获取，可能为 null）
 * @param x,y  原始鼠标坐标
 * @param margin 与视口边缘的安全距离
 */
export async function clampContextMenu(
  getEl: () => HTMLElement | null,
  x: number,
  y: number,
  margin = 8,
): Promise<void> {
  await nextTick()
  const el = getEl()
  if (!el) return
  const w = el.offsetWidth
  const h = el.offsetHeight
  const left = Math.max(margin, Math.min(x, window.innerWidth - w - margin))
  const top = Math.max(margin, Math.min(y, window.innerHeight - h - margin))
  el.style.left = `${left}px`
  el.style.top = `${top}px`
}
