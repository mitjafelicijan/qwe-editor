; Inherit C queries
(function_declarator (identifier) @function)
(call_expression (identifier) @function)
(parameter_declaration (primitive_type) @type)
(parameter_declaration (identifier) @variable)
(declaration (primitive_type) @type)
(declaration (identifier) @variable)
(string_literal) @string
(number_literal) @number
(char_literal) @string
(comment) @comment
(field_identifier) @property
((identifier) @boolean (#eq? @boolean "true"))
((identifier) @boolean (#eq? @boolean "false"))
(null) @null
((identifier) @keyword (#match? @keyword "^(static_cast|dynamic_cast|const_cast|reinterpret_cast|friend|inline|decltype|explicit|export|mutable|asm|auto|bool|nullptr)$"))

; C++ specific
(template_declaration) @type
(virtual) @keyword
(this) @keyword
(class_specifier name: (type_identifier) @type)
(namespace_definition name: (namespace_identifier) @type)
(using_declaration (qualified_identifier) @type)
(destructor_name) @function
(function_declarator (field_identifier) @function)
(function_definition declarator: (function_declarator declarator: (field_identifier) @function))

[
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
  "struct"
  "enum"
  "union"
  "typedef"
  "extern"
  "static"
  "const"
  "signed"
  "unsigned"
  "volatile"
  
  ; C++ keywords
  "new"
  "delete"
  "operator"
  "throw"
  "try"
  "catch"
  "class"
  "constexpr"
  "template"
  "typename"
  "using"
  "namespace"
  "public"
  "private"
  "protected"

  "#include"
  "#define"
  "#ifdef"
  "#ifndef"
  "#endif"
] @keyword