@file:JvmName("ImplUtil")
package com.vc2402.sdfplugin.psi.impl

import com.vc2402.sdfplugin.psi.SDFField
import com.vc2402.sdfplugin.psi.SDFParam
import com.vc2402.sdfplugin.psi.Types

fun getType(element: SDFParam): String? {
    return element.node.findChildByType(Types.TYPE)?.text
}

fun getType(element: SDFField): String? {
    return element.node.findChildByType(Types.TYPE)?.text
}
