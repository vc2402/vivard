package com.vc2402.sdfplugin.psi

// Copyright 2000-2022 JetBrains s.r.o. and other contributors. Use of this source code is governed by the Apache 2.0 license that can be found in the LICENSE file.

import com.intellij.extapi.psi.PsiFileBase
import com.intellij.openapi.fileTypes.FileType
import com.intellij.psi.FileViewProvider
import com.intellij.psi.util.elementType
import com.vc2402.sdfplugin.SDFFileType
import com.vc2402.sdfplugin.SDFLanguage
import com.vc2402.sdfplugin.TypeDescription


class SDFPsiFile(viewProvider: FileViewProvider) :
    PsiFileBase(viewProvider, SDFLanguage.INSTANCE) {
    private var pckg: String? = null
    private var types: MutableList<TypeDescription>? = null

    override fun getFileType(): FileType {
        return SDFFileType.INSTANCE
    }

    override fun toString(): String {
        return "SDF File"
    }

    fun getPackage(): String {
        if(pckg == null) {
            var el = firstChild
            while (el != null) {
                if (el.elementType == Types.PACKAGE_DECLARATION) {
                    var pe = el.firstChild
                    while (pe != null && pe.elementType != Types.IDENTIFIER)
                        pe = pe.nextSibling
                    if (pe != null)
                        pckg = pe.text
                    break
                }
                if (el.elementType != Types.COMMENTS)
                    break
                el = el.nextSibling
            }
        }
        if(pckg == null)
            pckg = ""
        return pckg!!
    }

    fun getTypes(): List<TypeDescription> {
        if(types == null) {
            types = mutableListOf()
            var el = firstChild
            while(el != null) {
                if(el.elementType == Types.PACKAGE_DECLARATION) {
                    var pe = el.firstChild
                    while(pe != null && pe.elementType != Types.IDENTIFIER)
                        pe = pe.nextSibling
                    if(pe != null)
                        pckg = pe.text
                }
                if(el.elementType == Types.DECLARATIONS) {
                    el = el.firstChild
                    continue
                }
                if(el.elementType == Types.DECLARATION &&
                    el.firstChild != null &&
                    el.firstChild.elementType == Types.TYPE_DECLARATION) {
                    lateinit var name: String
                    var extendable = false
                    var embeddable = false
                    var dictionary = false
                    var de = el.firstChild.firstChild
                    while(de != null) {
                        if(de.elementType == Types.TYPE_MODIFIERS) {
                            var tm = de.firstChild
                            while(tm != null) {
                                if(tm.firstChild != null && tm.firstChild.elementType == Types.TYPEMODIFIER) {
                                    if(tm.text == "extendable")
                                        extendable = true
                                    else if(tm.text == "dictionary")
                                        dictionary = true
                                    else if(tm.text == "embeddable")
                                        embeddable = true
                                }
                                tm = tm.nextSibling
                            }
                        }
                        if(de.elementType == Types.KW_TYPE) {
                            var ee = de.nextSibling
                            while(ee != null) {
                                if (ee.elementType == Types.IDENTIFIER) {
                                    name = ee.text
                                    break
                                }
                                ee = ee.nextSibling
                            }
                            break
                        }
                        de = de.nextSibling
                    }
                    types!!.add(TypeDescription(pckg?:"", name, extendable, embeddable, dictionary))
                }
                el = el.nextSibling
            }
        }
        return types!!
    }

    override fun clearCaches() {
        super.clearCaches()
        pckg = null
        types = null
    }
}