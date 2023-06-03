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
HOOKNAME = "create" | "change" | "changed" | "start" | "resolve" | "set" | {time_hook}
time_hook = "time="{STRING_VALUE}
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

//%state WAITING_ID
%%

<YYINITIAL> {COMMENT_LINE}                           { yybegin(YYINITIAL); return Types.COMMENT_LINE; }

<YYINITIAL> {DUMMYIDENTIFIER}                           { yybegin(YYINITIAL); return Types.DUMMYIDENTIFIER; }

<YYINITIAL> {STATEMENT_END}                           { yybegin(YYINITIAL); return Types.STATEMENT_END; }

<YYINITIAL> {PACKAGE}                           { yybegin(YYINITIAL); return Types.PACKAGE; }

<YYINITIAL> {KW_TYPE}                                { yybegin(YYINITIAL); return Types.KW_TYPE; }

<YYINITIAL> {EXTENDS}                                { yybegin(YYINITIAL); return Types.EXTENDS; }

<YYINITIAL> {META}                                { yybegin(YYINITIAL); return Types.META; }

<YYINITIAL> {MAP}                                { yybegin(YYINITIAL); return Types.MAP; }

<YYINITIAL> {INT}                                { yybegin(YYINITIAL); return Types.INT; }

<YYINITIAL> {STRING}                                { yybegin(YYINITIAL); return Types.STRING; }

<YYINITIAL> {FLOAT}                                { yybegin(YYINITIAL); return Types.FLOAT; }

<YYINITIAL> {BOOL}                                { yybegin(YYINITIAL); return Types.BOOL; }

<YYINITIAL> {DATE}                                { yybegin(YYINITIAL); return Types.DATE; }

<YYINITIAL> {META_LINE}                                { yybegin(YYINITIAL); return Types.META_LINE; }

<YYINITIAL> {TYPEMODIFIER}                                     { yybegin(YYINITIAL); return Types.TYPEMODIFIER; }

<YYINITIAL> {ATTRMODIFIER}                                     { yybegin(YYINITIAL); return Types.ATTRMODIFIER; }

<YYINITIAL> {MORE}                                     { yybegin(YYINITIAL); return Types.MORE; }

<YYINITIAL> {NOTNULL}                                     { yybegin(YYINITIAL); return Types.NOTNULL; }

<YYINITIAL> {QUALIFIEDNAME}                                     { yybegin(YYINITIAL); return Types.QUALIFIEDNAME; }

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

<YYINITIAL> {BR_CLOSE}                                     { yybegin(YYINITIAL); return Types.BR_CLOSE; }

<YYINITIAL> {BRACKETOPEN}                                     { yybegin(YYINITIAL); return Types.BRACKETOPEN; }

<YYINITIAL> {BRACKETCLOSE}                                     { yybegin(YYINITIAL); return Types.BRACKETCLOSE; }

<YYINITIAL> {BRACESOPEN}                                     { yybegin(YYINITIAL); return Types.BRACESOPEN; }

<YYINITIAL> {BRACESCLOSE}                                     { yybegin(YYINITIAL); return Types.BRACESCLOSE; }

<YYINITIAL> {SEMI}                                     { yybegin(YYINITIAL); return Types.SEMI; }

<YYINITIAL> {EQUAL}                                     { yybegin(YYINITIAL); return Types.EQUAL; }

<YYINITIAL> {COMMA}                                     { yybegin(YYINITIAL); return Types.COMMA; }

<YYINITIAL> {MODIFIEROPEN}                                     { yybegin(YYINITIAL); return Types.MODIFIEROPEN; }

<YYINITIAL> {MODIFIERCLOSE}                                     { yybegin(YYINITIAL); return Types.MODIFIERCLOSE; }

({CRLF}|{WHITE_SPACE})+                                     { yybegin(YYINITIAL); return TokenType.WHITE_SPACE; }

[^]                                                         { return TokenType.BAD_CHARACTER; }