package com.vc2402.sdfplugin

import com.intellij.codeInsight.CodeInsightSettings
import com.intellij.codeInsight.editorActions.TypedHandlerDelegate
import com.intellij.codeInsight.editorActions.TypedHandlerUtil
import com.intellij.openapi.editor.Editor
import com.intellij.openapi.editor.highlighter.HighlighterIterator
import com.intellij.openapi.fileTypes.FileType
import com.intellij.openapi.project.Project
import com.intellij.psi.PsiFile
import com.intellij.psi.TokenType
import com.intellij.psi.tree.IElementType
import com.intellij.psi.tree.TokenSet
import com.vc2402.sdfplugin.psi.SDFPsiFile
import com.vc2402.sdfplugin.psi.Types

class SDFTypeHandleDelegate : TypedHandlerDelegate() {

    private var ltTyped = false

    override fun beforeCharTyped(c: Char, project: Project, editor: Editor, file: PsiFile, fileType: FileType): Result {
        if (file !is SDFPsiFile) return Result.CONTINUE

        ltTyped = c == '<' &&
//                TypedHandlerUtil.isAfterClassLikeIdentifierOrDot(
//            editor.caretModel.offset,
//            editor,
//            Types.IDENTIFIER,
//            Types.IDENTIFIER,
//            true
//        ) &&
                CodeInsightSettings.getInstance().AUTOINSERT_PAIR_BRACKET

        if (c == '>' && TypedHandlerUtil.handleGenericGT(
                editor,
                Types.MODIFIEROPEN,
                Types.MODIFIERCLOSE,
                INVALID_INSIDE_REFERENCE
            ) && CodeInsightSettings.getInstance().AUTOINSERT_PAIR_BRACKET
        ) return Result.STOP

        return Result.CONTINUE
    }

    override fun charTyped(c: Char, project: Project, editor: Editor, file: PsiFile): Result {
        if (file !is SDFPsiFile || !ltTyped) return Result.CONTINUE

        ltTyped = false
        TypedHandlerUtil.handleAfterGenericLT(
            editor,
            Types.MODIFIEROPEN,
            Types.MODIFIERCLOSE,
            INVALID_INSIDE_REFERENCE
        )
        if(checkIfSemicolonRequired(editor.caretModel.offset, editor))
            editor.document.insertString(editor.caretModel.offset  +1, ";")
        return Result.STOP
    }

    companion object {
        val INVALID_INSIDE_REFERENCE = TokenSet.create(
//            Types.BR_OPEN,
//            Types.BR_CLOSE,
//            Types.BRACESOPEN,
//            Types.BRACESCLOSE
        )
    }
    fun checkIfSemicolonRequired(offset: Int,
                                 editor: Editor):Boolean {
        var iterator: HighlighterIterator = editor.highlighter.createIterator(offset)
        var currPos = offset
        if (iterator.atEnd()) return false
        var pos = offset;
        lateinit var tokenType:IElementType
        iterator.advance()
        while(!iterator.atEnd()) {
            tokenType = iterator.tokenType
            if (tokenType == Types.STATEMENT_END)
                return false
            if(tokenType != TokenType.WHITE_SPACE)
                break
            iterator.advance()
            currPos ++
        }
        iterator.retreat()
        currPos --
        if (offset != iterator.end && iterator.start > 0) iterator.retreat()
        while(currPos > 0) {
            tokenType = iterator.tokenType
            if(tokenType == Types.SEMI)
                return true
            if(tokenType == Types.BRACESCLOSE)
                return false
            iterator.retreat()
            currPos --
        }

        return false
    }
}