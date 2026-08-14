!ifndef SHELL_CONTEXT
    !if ${REQUEST_EXECUTION_LEVEL} == "admin"
        !define SHELL_CONTEXT HKLM
    !else
        !define SHELL_CONTEXT HKCU
    !endif
!endif

## 文件关联宏（扩展 notepad-- 风格）
##
## 把 EasyText 注册到一组常见文本后缀的"打开方式"列表（不抢已有默认）。
##
## 写入位置（与 SHELL_CONTEXT 一致）：
##   HKCU/HKLM\Software\Classes\.${EXT}\OpenWithProgids\EasyText.text = ""
##   HKCU/HKLM\Software\Classes\EasyText.text                       = "EasyText 文本文件"
##   HKCU/HKLM\Software\Classes\EasyText.text\DefaultIcon           = "$INSTDIR\${PRODUCT_EXECUTABLE}",0
##   HKCU/HKLM\Software\Classes\EasyText.text\shell\open\command    = '"$INSTDIR\${PRODUCT_EXECUTABLE}" "%1"'
##
## 卸载时反向写回"打开方式"键值清理，保证用户移除 EasyText 后其它程序仍能正常"打开方式"显示。
##
## 注意：写 OpenWithProgids 是最低侵入方式——不会动 .${EXT} 主键指向，
## 已绑定的默认程序（如 VSCode 绑 .json）保持不变，EasyText 仅出现在右键"打开方式"菜单。

!macro EASY_TEXT_REGISTER_EXTENSION EXT
    ; 让 EasyText 出现在「打开方式」列表
    WriteRegStr SHELL_CONTEXT "Software\Classes\.${EXT}\OpenWithProgids" "EasyText.text" ""

    ; 在用户主动选了 "始终用 EasyText" 之后，HKCR\.${EXT} 主键会指向 EasyText.text；
    ; 此处仅写入 ProgID 定义（Description / Icon / open command），不覆盖 HKCR\.${EXT} 主键。
    WriteRegStr SHELL_CONTEXT "Software\Classes\EasyText.text" "" 'EasyText 文本文件'
    WriteRegStr SHELL_CONTEXT "Software\Classes\EasyText.text\DefaultIcon" "" '"$INSTDIR\${PRODUCT_EXECUTABLE}",0'
    WriteRegStr SHELL_CONTEXT "Software\Classes\EasyText.text\shell" "" "open"
    WriteRegStr SHELL_CONTEXT "Software\Classes\EasyText.text\shell\open" "" "&Open with EasyText"
    WriteRegStr SHELL_CONTEXT "Software\Classes\EasyText.text\shell\open\command" "" '"$INSTDIR\${PRODUCT_EXECUTABLE}" "%1"'
!macroend

!macro EASY_TEXT_UNREGISTER_EXTENSION EXT
    DeleteRegValue SHELL_CONTEXT "Software\Classes\.${EXT}\OpenWithProgids" "EasyText.text"
!macroend

; 集中声明扩展名清单；新增后缀在这里加一行 EASY_TEXT_REGISTER_EXTENSION 调用即可。
;
; 仅注册 OpenWithProgids（出现在「打开方式」列表，不抢已有默认）；
; 卸载时反向清理（见 project.nsi uninstall 段）。
!macro EASY_TEXT_REGISTER_ALL
    !insertmacro EASY_TEXT_REGISTER_EXTENSION "txt"
    !insertmacro EASY_TEXT_REGISTER_EXTENSION "log"
    !insertmacro EASY_TEXT_REGISTER_EXTENSION "md"
    !insertmacro EASY_TEXT_REGISTER_EXTENSION "markdown"
    !insertmacro EASY_TEXT_REGISTER_EXTENSION "json"
    !insertmacro EASY_TEXT_REGISTER_EXTENSION "yaml"
    !insertmacro EASY_TEXT_REGISTER_EXTENSION "yml"
    !insertmacro EASY_TEXT_REGISTER_EXTENSION "xml"
    !insertmacro EASY_TEXT_REGISTER_EXTENSION "ini"
    !insertmacro EASY_TEXT_REGISTER_EXTENSION "csv"
    !insertmacro EASY_TEXT_REGISTER_EXTENSION "tsv"
    !insertmacro EASY_TEXT_REGISTER_EXTENSION "conf"
    !insertmacro EASY_TEXT_REGISTER_EXTENSION "cfg"
    !insertmacro EASY_TEXT_REGISTER_EXTENSION "properties"
    !insertmacro EASY_TEXT_REGISTER_EXTENSION "text"
!macroend

!macro EASY_TEXT_UNREGISTER_ALL
    !insertmacro EASY_TEXT_UNREGISTER_EXTENSION "txt"
    !insertmacro EASY_TEXT_UNREGISTER_EXTENSION "log"
    !insertmacro EASY_TEXT_UNREGISTER_EXTENSION "md"
    !insertmacro EASY_TEXT_UNREGISTER_EXTENSION "markdown"
    !insertmacro EASY_TEXT_UNREGISTER_EXTENSION "json"
    !insertmacro EASY_TEXT_UNREGISTER_EXTENSION "yaml"
    !insertmacro EASY_TEXT_UNREGISTER_EXTENSION "yml"
    !insertmacro EASY_TEXT_UNREGISTER_EXTENSION "xml"
    !insertmacro EASY_TEXT_UNREGISTER_EXTENSION "ini"
    !insertmacro EASY_TEXT_UNREGISTER_EXTENSION "csv"
    !insertmacro EASY_TEXT_UNREGISTER_EXTENSION "tsv"
    !insertmacro EASY_TEXT_UNREGISTER_EXTENSION "conf"
    !insertmacro EASY_TEXT_UNREGISTER_EXTENSION "cfg"
    !insertmacro EASY_TEXT_UNREGISTER_EXTENSION "properties"
    !insertmacro EASY_TEXT_UNREGISTER_EXTENSION "text"
!macroend
