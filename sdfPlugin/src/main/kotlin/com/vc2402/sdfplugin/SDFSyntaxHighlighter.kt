package com.vc2402.sdfplugin

import com.intellij.lexer.Lexer
import com.intellij.openapi.editor.DefaultLanguageHighlighterColors
import com.intellij.openapi.editor.HighlighterColors
import com.intellij.openapi.editor.colors.TextAttributesKey
import com.intellij.openapi.fileTypes.SyntaxHighlighterBase
import com.intellij.psi.TokenType
import com.intellij.psi.tree.IElementType
import com.vc2402.sdfplugin.psi.Types

class SDFSyntaxHighlighter : SyntaxHighlighterBase() {
    override fun getHighlightingLexer(): Lexer {
        return LexerAdapter()
    }

    override fun getTokenHighlights(tokenType: IElementType): Array<out TextAttributesKey?> {
        when(tokenType) {
            Types.KW_TYPE, Types.KW_ENUM, Types.PACKAGE, Types.EXTENDS, Types.MAP -> return KEY_KEYS
            Types.INT, Types.FLOAT, Types.STRING, Types.BOOL, Types.DATE -> return TYPE
            Types.COMMENT_LINE -> return COMMENT_KEYS
            Types.IDENTIFIER -> return IDENTIFIER_KEYS
            Types.STRING_VALUE -> return STRING_VAL_KEYS
            Types.NUMBER_VALUE -> return NUMBER_VAL_KEYS
            Types.BOOL_VALUE -> return BOOL_VALUE_KEYS
            Types.ATTRMODIFIER, Types.TYPEMODIFIER -> return MODIFIER_KEYS
            Types.ANNOTATIONTAG -> return ANNOTATION_KEYS
            Types.HOOKTAG -> return HOOK_KEYS
        }
        return if (tokenType == TokenType.BAD_CHARACTER) {
            BAD_CHAR_KEYS
        } else EMPTY_KEYS
    }

    companion object {
        val IDENT = TextAttributesKey.createTextAttributesKey(
            "SIMPLE_SEPARATOR",
            DefaultLanguageHighlighterColors.IDENTIFIER
        )
        val KEY = TextAttributesKey.createTextAttributesKey("SDF_KEYWORD", DefaultLanguageHighlighterColors.KEYWORD)
        val STR_VAL = TextAttributesKey.createTextAttributesKey("SDF_STRING_VALUE", DefaultLanguageHighlighterColors.STRING)
        val NUMB_VAL = TextAttributesKey.createTextAttributesKey("SDF_NUMBER_VALUE", DefaultLanguageHighlighterColors.NUMBER)
        val BOOL_VAL = TextAttributesKey.createTextAttributesKey("SDF_BOOL_VALUE", DefaultLanguageHighlighterColors.NUMBER)
        val SIMPLE_TYPE = TextAttributesKey.createTextAttributesKey("SDF_TYPE", DefaultLanguageHighlighterColors.PREDEFINED_SYMBOL)
        val COMMENT =
            TextAttributesKey.createTextAttributesKey("SDF_COMMENT", DefaultLanguageHighlighterColors.LINE_COMMENT)
        val BAD_CHARACTER =
            TextAttributesKey.createTextAttributesKey("SDF_BAD_CHARACTER", HighlighterColors.BAD_CHARACTER)
        val MODIFIER = TextAttributesKey.createTextAttributesKey("SDF_MODIFIER", DefaultLanguageHighlighterColors.LABEL)
        val HOOK = TextAttributesKey.createTextAttributesKey("SDF_HOOK", DefaultLanguageHighlighterColors.INTERFACE_NAME)
        val ANNOTATION = TextAttributesKey.createTextAttributesKey("SDF_ANNOTATION", DefaultLanguageHighlighterColors.MARKUP_ATTRIBUTE)
        private val BAD_CHAR_KEYS = arrayOf(BAD_CHARACTER)
        private val IDENTIFIER_KEYS = arrayOf(IDENT)
        private val KEY_KEYS = arrayOf(KEY)
        private val TYPE = arrayOf(SIMPLE_TYPE)
        private val COMMENT_KEYS = arrayOf(COMMENT)
        private val STRING_VAL_KEYS = arrayOf(STR_VAL)
        private val NUMBER_VAL_KEYS = arrayOf(NUMB_VAL)
        private val BOOL_VALUE_KEYS = arrayOf(BOOL_VAL)
        private val MODIFIER_KEYS = arrayOf(MODIFIER)
        private val HOOK_KEYS = arrayOf(HOOK)
        private val ANNOTATION_KEYS = arrayOf(ANNOTATION)
        private val EMPTY_KEYS = arrayOfNulls<TextAttributesKey>(0)
    }
}