(function_declaration name: (identifier) @function)
(method_definition name: (property_identifier) @function)
(call_expression function: (identifier) @function)
(call_expression function: (member_expression property: (property_identifier) @function))

(type_identifier) @type
(predefined_type) @type

(string) @string
(number) @number
(comment) @comment
(regex) @string
(template_string) @string

(true) @boolean
(false) @boolean
(null) @null

(property_signature name: (property_identifier) @property)
(public_field_definition name: (property_identifier) @property)

[
  "function"
  "return"
  "if"
  "else"
  "for"
  "while"
  "do"
  "switch"
  "case"
  "default"
  "break"
  "continue"
  "var"
  "let"
  "const"
  "try"
  "catch"
  "finally"
  "class"
  "extends"
  "implements"
  "import"
  "export"
  "from"
  "async"
  "await"
  "new"
  "interface"
  "type"
  "enum"
  "public"
  "private"
  "protected"
  "readonly"
  "declare"
  "module"
  "namespace"
  "abstract"
  "as"
  "keyof"
  "typeof"
  "instanceof"
  "void"
  "debugger"
  "yield"
] @keyword

(this) @keyword
(super) @keyword

(jsx_opening_element name: (_) @function)
(jsx_closing_element name: (_) @function)
(jsx_self_closing_element name: (_) @function)
(jsx_attribute (property_identifier) @property)
