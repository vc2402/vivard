package com.vc2402.sdfplugin

import com.intellij.lang.BracePair
import com.intellij.lang.PairedBraceMatcher
import com.intellij.psi.PsiFile
import com.intellij.psi.tree.IElementType
import com.vc2402.sdfplugin.psi.Types
import org.jetbrains.annotations.NotNull
import org.jetbrains.annotations.Nullable


class SDFBraceMatcher: PairedBraceMatcher  {
    private val PAIRS = arrayOf(
        BracePair(Types.BRACESOPEN, Types.BRACESCLOSE, false),
        BracePair(Types.BRACKETOPEN, Types.BRACKETCLOSE, true),
        BracePair(Types.BR_OPEN, Types.BR_CLOSE, false),
        BracePair(Types.MODIFIEROPEN, Types.MODIFIERCLOSE, false)
    )


    override fun getPairs(): Array<BracePair> {
        return PAIRS
    }

    override fun isPairedBracesAllowedBeforeType(
        @NotNull lbraceType: IElementType,
        @Nullable contextType: IElementType?
    ): Boolean {
        return true
    }


    override fun getCodeConstructStart(file: PsiFile?, openingBraceOffset: Int): Int {
        return openingBraceOffset
    }
}