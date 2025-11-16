%{
package sqlite

import "strings"
%}

%union{
    str string
    strList []string
    column *columnDef
    elements []tableElement
    element tableElement
    constraint *constraintDef
    col columnConstraint
    colcons []columnConstraint
    boolVal bool
    qname qualifiedName
    fkTarget fkTarget
    fkActions fkActions
}

%type <boolVal> opt_if_not_exists opt_unique opt_temp
%type <elements> table_element_list
%type <element> table_element
%type <column> column_def
%type <colcons> column_constraint_list
%type <col> column_constraint
%type <constraint> table_constraint
%type <str> identifier type_name type_params opt_constraint_name opt_index_name default_value signed_number
%type <strList> column_list indexed_column_list opt_reference_columns
%type <str> indexed_column
%type <qname> qualified_name
%type <fkTarget> reference_target
%type <fkActions> opt_fk_actions fk_action
%type <str> fk_action_value

%token CREATE_KW TABLE_KW TEMP_KW IF_KW NOT_KW EXISTS_KW PRIMARY_KW KEY_KW UNIQUE_KW CONSTRAINT_KW FOREIGN_KW REFERENCES_KW DEFAULT_KW AUTOINCREMENT_KW
%token ON_KW DELETE_KW UPDATE_KW SET_KW CASCADE_KW RESTRICT_KW NO_KW ACTION_KW WITHOUT_KW ROWID_KW INDEX_KW
%token TRUE_KW FALSE_KW NULL_KW CURRENT_TIMESTAMP_KW CURRENT_DATE_KW CURRENT_TIME_KW COLLATE_KW ASC_KW DESC_KW
%token <str> IDENT STRING NUMBER
%token LPAREN RPAREN COMMA SEMICOLON DOT PLUS MINUS
%%
ddl:
      /* empty */
    | ddl statement
    ;

statement:
      create_table_stmt opt_semicolon
    | create_index_stmt opt_semicolon
    ;

create_table_stmt:
      CREATE_KW opt_temp TABLE_KW opt_if_not_exists qualified_name LPAREN table_element_list RPAREN opt_table_suffix {
          lex := sqlitelex.(*lexer)
          tb := newTableBuilder($5.schema, $5.name)
          for _, elem := range $7 {
              if elem.column != nil {
                  tb.addColumn(elem.column)
              }
              if elem.constraint != nil {
                  tb.applyConstraint(elem.constraint)
              }
          }
          lex.addTable(tb)
      }
    ;

opt_temp:
      /* empty */ { $$ = false }
    | TEMP_KW { $$ = true }
    ;

opt_table_suffix:
      /* empty */
    | WITHOUT_KW ROWID_KW
    ;

table_element_list:
      table_element { $$ = []tableElement{$1} }
    | table_element_list COMMA table_element { $$ = append($1, $3) }
    | table_element_list COMMA { $$ = $1 }
    ;

table_element:
      column_def { $$ = tableElement{column: $1} }
    | table_constraint { $$ = tableElement{constraint: $1} }
    ;

column_def:
      identifier type_name column_constraint_list {
          col := &columnDef{name: $1, dbType: $2}
          applyColumnConstraints(col, $3)
          $$ = col
      }
    ;

type_name:
      identifier { $$ = $1 }
    | type_name identifier { $$ = $1 + " " + $2 }
    | type_name LPAREN type_params RPAREN { $$ = $1 + "(" + $3 + ")" }
    ;

type_params:
      NUMBER { $$ = $1 }
    | type_params COMMA NUMBER { $$ = $1 + "," + $3 }
    ;

column_constraint_list:
      /* empty */ { $$ = nil }
    | column_constraint_list column_constraint { $$ = append($1, $2) }
    ;

column_constraint:
      NOT_KW NULL_KW { $$ = columnConstraint{kind: columnConstraintNotNull} }
    | NULL_KW { $$ = columnConstraint{kind: columnConstraintNull} }
    | PRIMARY_KW KEY_KW { $$ = columnConstraint{kind: columnConstraintPrimaryKey} }
    | UNIQUE_KW { $$ = columnConstraint{kind: columnConstraintUnique} }
    | DEFAULT_KW default_value { $$ = columnConstraint{kind: columnConstraintDefault, value: $2} }
    | REFERENCES_KW reference_target {
          ref := $2
          $$ = columnConstraint{kind: columnConstraintForeignKey, fk: &ref}
      }
    | AUTOINCREMENT_KW { $$ = columnConstraint{kind: columnConstraintPrimaryKey} }
    ;

table_constraint:
      opt_constraint_name PRIMARY_KW KEY_KW LPAREN column_list RPAREN {
          $$ = &constraintDef{kind: constraintPrimaryKey, name: $1, columns: $5}
      }
    | opt_constraint_name UNIQUE_KW opt_index_name LPAREN column_list RPAREN {
          name := $1
          if name == "" {
              name = $3
          }
          $$ = &constraintDef{kind: constraintUnique, name: name, columns: $5}
      }
    | opt_constraint_name FOREIGN_KW opt_key_kw LPAREN column_list RPAREN REFERENCES_KW reference_target {
          $$ = &constraintDef{
              kind:    constraintForeignKey,
              name:    $1,
              columns: $5,
              fk:      $8,
          }
      }
    ;

opt_key_kw:
      /* empty */
    | KEY_KW
    ;

opt_index_name:
      /* empty */ { $$ = "" }
    | identifier { $$ = $1 }
    | KEY_KW identifier { $$ = $2 }
    | INDEX_KW identifier { $$ = $2 }
    ;

reference_target:
      qualified_name opt_reference_columns opt_fk_actions {
          $$ = fkTarget{
              table:   $1,
              columns: $2,
              actions: $3,
          }
      }
    ;

opt_reference_columns:
      /* empty */ { $$ = nil }
    | LPAREN column_list RPAREN { $$ = $2 }
    ;

opt_fk_actions:
      /* empty */ { $$ = fkActions{} }
    | opt_fk_actions fk_action {
          $$ = $1
          if $2.onDelete != "" {
              $$.onDelete = $2.onDelete
          }
          if $2.onUpdate != "" {
              $$.onUpdate = $2.onUpdate
          }
      }
    ;

fk_action:
      ON_KW DELETE_KW fk_action_value { $$ = fkActions{onDelete: $3} }
    | ON_KW UPDATE_KW fk_action_value { $$ = fkActions{onUpdate: $3} }
    ;

fk_action_value:
      CASCADE_KW { $$ = "CASCADE" }
    | RESTRICT_KW { $$ = "RESTRICT" }
    | NO_KW ACTION_KW { $$ = "NO ACTION" }
    | SET_KW NULL_KW { $$ = "SET NULL" }
    | SET_KW DEFAULT_KW { $$ = "SET DEFAULT" }
    ;

column_list:
      identifier { $$ = []string{$1} }
    | column_list COMMA identifier { $$ = append($1, $3) }
    ;

identifier:
      IDENT { $$ = $1 }
    ;

opt_constraint_name:
      /* empty */ { $$ = "" }
    | CONSTRAINT_KW identifier { $$ = $2 }
    ;

default_value:
      STRING { $$ = $1 }
    | NUMBER { $$ = $1 }
    | signed_number { $$ = $1 }
    | identifier { $$ = strings.ToUpper($1) }
    | NULL_KW { $$ = "NULL" }
    | TRUE_KW { $$ = "TRUE" }
    | FALSE_KW { $$ = "FALSE" }
    | CURRENT_TIMESTAMP_KW opt_empty_parens { $$ = "CURRENT_TIMESTAMP" }
    | CURRENT_DATE_KW opt_empty_parens { $$ = "CURRENT_DATE" }
    | CURRENT_TIME_KW opt_empty_parens { $$ = "CURRENT_TIME" }
    | LPAREN default_value RPAREN { $$ = $2 }
    ;

opt_empty_parens:
      /* empty */
    | LPAREN RPAREN
    ;

signed_number:
      PLUS NUMBER { $$ = "+" + $2 }
    | MINUS NUMBER { $$ = "-" + $2 }
    ;

qualified_name:
      identifier { $$ = qualifiedName{name: $1} }
    | identifier DOT identifier { $$ = qualifiedName{schema: $1, name: $3} }
    ;

create_index_stmt:
      CREATE_KW opt_unique INDEX_KW opt_if_not_exists identifier ON_KW qualified_name LPAREN indexed_column_list RPAREN {
          lex := sqlitelex.(*lexer)
          lex.addIndex(indexDef{
              schema:  $7.schema,
              table:   $7.name,
              name:    $5,
              columns: $9,
              unique:  $2,
          })
      }
    ;

opt_unique:
      /* empty */ { $$ = false }
    | UNIQUE_KW { $$ = true }
    ;

indexed_column_list:
      indexed_column { $$ = []string{$1} }
    | indexed_column_list COMMA indexed_column { $$ = append($1, $3) }
    ;

indexed_column:
      identifier opt_collate opt_sort_order { $$ = $1 }
    ;

opt_collate:
      /* empty */
    | COLLATE_KW identifier
    ;

opt_sort_order:
      /* empty */
    | ASC_KW
    | DESC_KW
    ;

opt_if_not_exists:
      /* empty */ { $$ = false }
    | IF_KW NOT_KW EXISTS_KW { $$ = true }
    ;

opt_semicolon:
      /* empty */
    | SEMICOLON
    ;
%%
