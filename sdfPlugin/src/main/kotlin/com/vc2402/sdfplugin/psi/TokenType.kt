package com.vc2402.sdfplugin.psi

import com.intellij.psi.tree.IElementType
import com.vc2402.sdfplugin.SDFLanguage
import org.jetbrains.annotations.NonNls


class TokenType(@NonNls debugName: String) :
    IElementType(debugName, SDFLanguage.INSTANCE) {
    override fun toString(): String {
        return "TokenType." + super.toString()
    }
}