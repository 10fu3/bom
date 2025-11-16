%{
package postgres

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

%type <boolVal> opt_if_not_exists opt_unique opt_only
%type <elements> table_element_list
%type <element> table_element
%type <column> column_def
%type <colcons> column_constraint_list
%type <col> column_constraint
%type <constraint> table_constraint
%type <str> identifier type_name type_params opt_constraint_name opt_index_name default_expr literal function_call qualified_identifier opt_table_suffix opt_expr_list expr_list opt_match_type
%type <strList> column_list opt_reference_columns
%type <qname> qualified_name
%type <fkTarget> reference_target
%type <fkActions> opt_fk_actions fk_action
%type <str> fk_action_value

%token CREATE_KW TABLE_KW INDEX_KW IF_KW NOT_KW EXISTS_KW PRIMARY_KW KEY_KW UNIQUE_KW CONSTRAINT_KW FOREIGN_KW REFERENCES_KW DEFAULT_KW ON_KW DELETE_KW UPDATE_KW SET_KW CASCADE_KW RESTRICT_KW NO_KW ACTION_KW NULL_KW TRUE_KW FALSE_KW CURRENT_TIMESTAMP_KW CURRENT_DATE_KW CURRENT_TIME_KW
%token ALTER_KW ADD_KW COLUMN_KW DROP_KW ONLY_KW TABLESPACE_KW WITH_KW CHECK_KW MATCH_KW DEFERRABLE_KW INITIALLY_KW DEFERRED_KW IMMEDIATE_KW GENERATED_KW ALWAYS_KW BY_KW AS_KW IDENTITY_KW USING_KW INCLUDE_KW COMMENT_KW IS_KW COLLATE_KW
%token <str> IDENT STRING NUMBER
%token LPAREN RPAREN COMMA SEMICOLON DOT PLUS MINUS STAR SLASH TYPECAST LBRACKET RBRACKET EQUALS

%left PLUS MINUS
%left STAR SLASH
%right UMINUS UPLUS
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
      CREATE_KW table_modifier_opt TABLE_KW opt_if_not_exists qualified_name LPAREN table_element_list RPAREN opt_table_suffix {
          lex := postgreslex.(*lexer)
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

table_modifier_opt:
      /* empty */
    | IDENT
    ;

opt_table_suffix:
      /* empty */ { $$ = "" }
    | WITH_KW LPAREN with_option_list RPAREN { $$ = "" }
    | TABLESPACE_KW identifier { $$ = $2 }
    ;

with_option_list:
      with_option
    | with_option_list COMMA with_option
    ;

with_option:
      identifier
    | identifier EQUALS identifier
    | identifier EQUALS NUMBER
    | identifier EQUALS STRING
    ;

opt_if_not_exists:
      /* empty */ { $$ = false }
    | IF_KW NOT_KW EXISTS_KW { $$ = true }
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
    | type_name LBRACKET RBRACKET { $$ = $1 + "[]" }
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
    | DEFAULT_KW default_expr { $$ = columnConstraint{kind: columnConstraintDefault, value: $2} }
    | REFERENCES_KW reference_target opt_match_type opt_fk_actions opt_deferrable {
          ref := $2
          _ = $3
          ref.actions = $4
          $$ = columnConstraint{kind: columnConstraintForeignKey, fk: &ref}
      }
    | CHECK_KW balanced_parens { $$ = columnConstraint{kind: columnConstraintNoop} }
    | COLLATE_KW identifier { $$ = columnConstraint{kind: columnConstraintNoop} }
    | GENERATED_KW generated_clause { $$ = columnConstraint{kind: columnConstraintNoop} }
    ;

generated_clause:
      ALWAYS_KW AS_KW IDENTITY_KW
    | BY_KW DEFAULT_KW AS_KW IDENTITY_KW
    ;

balanced_parens:
      LPAREN balanced_tokens RPAREN
    ;

balanced_tokens:
      /* empty */
    | balanced_tokens default_expr
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
    | opt_constraint_name FOREIGN_KW opt_key_kw LPAREN column_list RPAREN REFERENCES_KW reference_target opt_match_type opt_fk_actions opt_deferrable {
          $$ = &constraintDef{
              kind:    constraintForeignKey,
              name:    $1,
              columns: $5,
              fk:      fkTarget{
                  table:   $8.table,
                  columns: $8.columns,
                  actions: $10,
              },
          }
      }
    | opt_constraint_name CHECK_KW balanced_parens { $$ = nil }
    ;

opt_key_kw:
      /* empty */
    | KEY_KW
    ;

opt_constraint_name:
      /* empty */ { $$ = "" }
    | CONSTRAINT_KW identifier { $$ = $2 }
    ;

opt_index_name:
      /* empty */ { $$ = "" }
    | identifier { $$ = $1 }
    ;

reference_target:
      opt_only qualified_name opt_reference_columns {
          _ = $1
          $$ = fkTarget{table: $2, columns: $3}
      }
    ;

opt_match_type:
      /* empty */ { $$ = "" }
    | MATCH_KW identifier { $$ = $2 }
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

opt_deferrable:
      /* empty */
    | DEFERRABLE_KW
    | NOT_KW DEFERRABLE_KW
    | INITIALLY_KW IMMEDIATE_KW
    | INITIALLY_KW DEFERRED_KW
    ;

column_list:
      identifier { $$ = []string{$1} }
    | column_list COMMA identifier { $$ = append($1, $3) }
    ;

qualified_name:
      identifier { $$ = qualifiedName{name: $1} }
    | identifier DOT identifier { $$ = qualifiedName{schema: $1, name: $3} }
    ;

qualified_identifier:
      identifier { $$ = $1 }
    | qualified_identifier DOT identifier { $$ = $1 + "." + $3 }
    ;

default_expr:
      literal { $$ = $1 }
    | qualified_identifier { $$ = $1 }
    | function_call { $$ = $1 }
    | LPAREN default_expr RPAREN { $$ = "(" + $2 + ")" }
    | default_expr TYPECAST qualified_identifier { $$ = $1 + "::" + $3 }
    | default_expr PLUS default_expr { $$ = $1 + " + " + $3 }
    | default_expr MINUS default_expr { $$ = $1 + " - " + $3 }
    | default_expr STAR default_expr { $$ = $1 + " * " + $3 }
    | default_expr SLASH default_expr { $$ = $1 + " / " + $3 }
    | MINUS default_expr %prec UMINUS { $$ = "-" + $2 }
    | PLUS default_expr %prec UPLUS { $$ = "+" + $2 }
    ;

literal:
      STRING { $$ = quoteLiteral($1) }
    | NUMBER { $$ = $1 }
    | TRUE_KW { $$ = "TRUE" }
    | FALSE_KW { $$ = "FALSE" }
    | NULL_KW { $$ = "NULL" }
    | CURRENT_TIMESTAMP_KW { $$ = "CURRENT_TIMESTAMP" }
    | CURRENT_DATE_KW { $$ = "CURRENT_DATE" }
    | CURRENT_TIME_KW { $$ = "CURRENT_TIME" }
    ;

function_call:
      qualified_identifier LPAREN opt_expr_list RPAREN {
          if $3 == "" {
              $$ = $1 + "()"
          } else {
              $$ = $1 + "(" + $3 + ")"
          }
      }
    ;

opt_expr_list:
      /* empty */ { $$ = "" }
    | expr_list { $$ = $1 }
    ;

expr_list:
      default_expr { $$ = $1 }
    | expr_list COMMA default_expr { $$ = $1 + ", " + $3 }
    ;

create_index_stmt:
      CREATE_KW opt_unique INDEX_KW identifier ON_KW opt_only qualified_name opt_index_method LPAREN column_list RPAREN opt_index_including opt_table_suffix {
          lex := postgreslex.(*lexer)
          lex.addIndex(indexDef{
              schema:  $7.schema,
              table:   $7.name,
              name:    $4,
              columns: $10,
              unique:  $2,
          })
      }
    ;

opt_unique:
      /* empty */ { $$ = false }
    | UNIQUE_KW { $$ = true }
    ;

opt_index_method:
      /* empty */
    | USING_KW identifier
    ;

opt_index_including:
      /* empty */
    | INCLUDE_KW LPAREN column_list RPAREN
    ;

opt_semicolon:
      /* empty */
    | SEMICOLON
    ;

identifier:
      IDENT { $$ = $1 }
    ;

opt_only:
      /* empty */ { $$ = false }
    | ONLY_KW { $$ = true }
    ;

%%
func quoteLiteral(in string) string {
    if strings.HasPrefix(in, "'") {
        return in
    }
    return "'" + strings.ReplaceAll(in, "'", "''") + "'"
}
