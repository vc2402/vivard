package com.vc2402.sdfplugin;

import com.intellij.lexer.FlexLexer;
import com.intellij.psi.tree.IElementType;
import com.vc2402.sdfplugin.psi.Types;
import com.intellij.psi.TokenType;

%%

%class SDFLexer
%implements FlexLexer
%unicode
%function advance
%type IElementType
%eof{  return;
%eof}

COMMENT_LINE = "/""/"[^\r\n]*
//DUMMYIDENTIFIER = "\u001f"
DUMMYIDENTIFIER = "DuM_Id"
META_LINE = [#\t][^\r\n]*
PACKAGE = "package"
KW_TYPE = "type"
EXTENDS = "extends"
META = "meta"
TYPEMODIFIER = "abstract" | "config" | "dictionary" | "transient" | "embeddable" | "singleton" | "extern" | "extendable"
ATTRMODIFIER = "id" | "auto" | "lookup" | "one-to-many" | "embedded" | "ref-embedded" | "calculated"
INT = "int"
FLOAT = "float"
STRING = "string"
BOOL = "bool"
DATE = "date"
MAP = "map"
AUTO = "auto"
BOOL_VALUE = "true" | "false"
MORE = "..."
STATEMENT_END = ";"
QUALIFIEDNAME = {IDENTIFIER} "." {IDENTIFIER}
IDENTIFIER = [a-zA-Z_][a-zA-Z0-9_]*
ANNOTATIONTAG = "$" {ANNOTATIONNAME} (":" {ANNOTATIONNAME})?
//ANNOTATIONSTART = "$"
ANNOTATIONNAME = [a-zA-Z_][a-zA-Z0-9_-]*
//HOOKTAG = "@"[a-zA-Z_][a-zA-Z0-9_-]*
HOOKTAG = "@" {HOOKNAME} (":" {IDENTIFIER})?
HOOKNAME = "create" | "change" | "changed" | "start" | "resolve" | "set" | "time"
//time_hook = "time="{STRING_VALUE}
STRING_VALUE = "\"" ([^\"\\] | "\\" {any} )* "\""
NUMBER_VALUE = ("." | {digit}) ("." | {digit})*
//PUNCT = "!"…"/" | ":"…"@" | "["…` + "\"`\"" + ` | "{"…"~" .
BR_OPEN = "("
BR_CLOSE = ")"
BRACKETOPEN = "["
BRACKETCLOSE = "]"
BRACESOPEN = "{"
BRACESCLOSE = "}"
MODIFIEROPEN = "<"
MODIFIERCLOSE = ">"
SEMI = ":"
EQUAL = "="
COMMA = ","
NOTNULL = "!"

CRLF=\R
WHITE_SPACE=[\ \n\t\f]

any = "\u0000"…"\uffff"
alpha = [a-zA-Z]
digit = [0-9]

%state WAITING_TYPE

%state WAITING_ATTR_MODIFIER

%%

<YYINITIAL> {COMMENT_LINE}                           { yybegin(YYINITIAL); return Types.COMMENT_LINE; }

<YYINITIAL, WAITING_TYPE> {STATEMENT_END}                           { yybegin(YYINITIAL); return Types.STATEMENT_END; }

<YYINITIAL> {PACKAGE}                           { yybegin(YYINITIAL); return Types.PACKAGE; }

<YYINITIAL> {KW_TYPE}                                { yybegin(YYINITIAL); return Types.KW_TYPE; }

<YYINITIAL> {EXTENDS}                                { yybegin(YYINITIAL); return Types.EXTENDS; }

<YYINITIAL> {META}                                { yybegin(YYINITIAL); return Types.META; }

<WAITING_TYPE> {
        {MAP}                                {return Types.MAP; }

        {STRING}                                { return Types.STRING; }

        {FLOAT}                                { return Types.FLOAT; }

        {BOOL}                                {  return Types.BOOL; }

        {AUTO}                                { return Types.AUTO; }

        {DATE}                                {  return Types.DATE; }

        {INT}                                {  return Types.INT; }

        {IDENTIFIER}                          {  return Types.IDENTIFIER; }

        {BRACKETOPEN}                        { return Types.BRACKETOPEN; }

        {BRACKETCLOSE}                        { return Types.BRACKETCLOSE; }

      {NOTNULL}                                     { return Types.NOTNULL; }
}



<WAITING_ATTR_MODIFIER> {
     {ATTRMODIFIER}                  { return Types.ATTRMODIFIER; }

     {DUMMYIDENTIFIER}              { return Types.DUMMYIDENTIFIER; }

     {ANNOTATIONTAG}                  { return Types.ANNOTATIONTAG; }

     {STRING_VALUE}                  { return Types.STRING_VALUE; }

     {BR_OPEN}                       {return Types.BR_OPEN; }

     {BR_CLOSE}                      { return Types.BR_CLOSE; }

      {EQUAL}                                     { return Types.EQUAL; }

      {NUMBER_VALUE}                                     {return Types.NUMBER_VALUE; }

      {BOOL_VALUE}                                     {return Types.BOOL_VALUE; }

     {MODIFIERCLOSE}                 { yybegin(YYINITIAL); return Types.MODIFIERCLOSE; }

     {IDENTIFIER}                  { return Types.IDENTIFIER; }

      {HOOKTAG}                                     {return Types.HOOKTAG; }
}

<YYINITIAL> {META_LINE}                                { yybegin(YYINITIAL); return Types.META_LINE; }

<YYINITIAL> {TYPEMODIFIER}                                     { yybegin(YYINITIAL); return Types.TYPEMODIFIER; }

//<YYINITIAL> {ATTRMODIFIER}                                     { yybegin(YYINITIAL); return Types.ATTRMODIFIER; }

<YYINITIAL> {MORE}                                     {return Types.MORE; }

<YYINITIAL> {NOTNULL}                                     { return Types.NOTNULL; }

<YYINITIAL, WAITING_TYPE> {QUALIFIEDNAME}                                     { return Types.QUALIFIEDNAME; }

//<WAITING_ID> {IDENTIFIER}                                     { yybegin(YYINITIAL); return Types.IDENTIFIER; }

<YYINITIAL> {STRING_VALUE}                                     { yybegin(YYINITIAL); return Types.STRING_VALUE; }

<YYINITIAL> {NUMBER_VALUE}                                     { yybegin(YYINITIAL); return Types.NUMBER_VALUE; }

<YYINITIAL> {BOOL_VALUE}                                     { yybegin(YYINITIAL); return Types.BOOL_VALUE; }

<YYINITIAL> {IDENTIFIER}                                     { yybegin(YYINITIAL); return Types.IDENTIFIER; }

<YYINITIAL> {ANNOTATIONTAG}                                     { yybegin(YYINITIAL); return Types.ANNOTATIONTAG; }

//<YYINITIAL> {ANNOTATIONNAME}                                     { yybegin(YYINITIAL); return Types.ANNOTATIONNAME; }

<YYINITIAL> {HOOKTAG}                                     { yybegin(YYINITIAL); return Types.HOOKTAG; }

<YYINITIAL> {WHITE_SPACE}                                     { yybegin(YYINITIAL); return TokenType.WHITE_SPACE; }

//<YYINITIAL> {PUNCT}                                     { yybegin(YYINITIAL); return Types.PUNCT; }

<YYINITIAL> {BR_OPEN}                                     { yybegin(YYINITIAL); return Types.BR_OPEN; }

<YYINITIAL, WAITING_TYPE> {BR_CLOSE}                                     { yybegin(YYINITIAL); return Types.BR_CLOSE; }

<YYINITIAL> {BRACKETOPEN}                                     { yybegin(YYINITIAL); return Types.BRACKETOPEN; }

<YYINITIAL> {BRACKETCLOSE}                                     { yybegin(YYINITIAL); return Types.BRACKETCLOSE; }

<YYINITIAL> {BRACESOPEN}                                     { yybegin(YYINITIAL); return Types.BRACESOPEN; }

<YYINITIAL> {BRACESCLOSE}                                     { yybegin(YYINITIAL); return Types.BRACESCLOSE; }

<YYINITIAL> {SEMI}                                     { yybegin(WAITING_TYPE); return Types.SEMI; }

<YYINITIAL> {EQUAL}                                     { return Types.EQUAL; }

{COMMA}                                     { yybegin(YYINITIAL); return Types.COMMA; }

<YYINITIAL, WAITING_TYPE> {MODIFIEROPEN}                                     { yybegin(WAITING_ATTR_MODIFIER); return Types.MODIFIEROPEN; }

//<YYINITIAL> {MODIFIERCLOSE}                                     { yybegin(YYINITIAL); return Types.MODIFIERCLOSE; }

({CRLF}|{WHITE_SPACE})+                                     { return TokenType.WHITE_SPACE; }

[^]                                                         { return TokenType.BAD_CHARACTER; }

{DUMMYIDENTIFIER}                           { return Types.DUMMYIDENTIFIER; }