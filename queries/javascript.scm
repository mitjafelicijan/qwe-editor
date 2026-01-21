(function_declaration name: (identifier) @function)
(method_definition name: (property_identifier) @function)
(call_expression function: (identifier) @function)
(call_expression function: (member_expression property: (property_identifier) @function))
(string) @string
(number) @number
(comment) @comment
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
  "import"
  "export"
  "default"
  "from"
  "async"
  "await"
  "new"
] @keyword
