/**
 * ext-keymap — 命令别名出口
 *
 * 历史上这里与 utils/commands.ts 各维护一份别名表且已漂移
 * （同一条 'line-insert-above' 两边映射到不同目标，全靠命令分发
 * 的宽松匹配兜底才没炸）。现统一收敛到 utils/commands.ts 的
 * CMD_ALIAS（单一数据源），这里仅按旧名字 re-export，
 * CodeEditor 的 `CMD_ALIASES` 引用保持不变。
 *
 * 说明：
 *  - keymap 装配由 CodeEditor.createEditor 内联完成（此前本文件的
 *    buildPrecKeymap 是从未被调用的死代码，已移除）；
 *  - 标点键（PunctuationKeymap）已被显式移除——CodeMirror 6
 *    已正确处理标点符号，旧实现拦截会与中文 IME 冲突。
 */
export { CMD_ALIAS as CMD_ALIASES } from '@/utils/commands'
