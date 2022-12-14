package com.vc2402.sdfplugin.psi

import com.intellij.psi.tree.IElementType
import com.vc2402.sdfplugin.SDFLanguage
import org.jetbrains.annotations.NonNls


class ElementType(@NonNls debugName: String) :
    IElementType(debugName, SDFLanguage.INSTANCE)