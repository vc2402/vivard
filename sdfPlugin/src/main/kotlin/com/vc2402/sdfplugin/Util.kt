package com.vc2402.sdfplugin

import com.google.common.collect.Lists
import com.intellij.find.editorHeaderActions.Embeddable
import com.intellij.openapi.project.Project
import com.intellij.openapi.vfs.VirtualFile
import com.intellij.psi.PsiComment
import com.intellij.psi.PsiElement
import com.intellij.psi.PsiManager
import com.intellij.psi.PsiWhiteSpace
import com.intellij.psi.search.FileTypeIndex
import com.intellij.psi.search.GlobalSearchScope
import com.intellij.psi.util.PsiTreeUtil
import com.intellij.psi.util.elementType
import com.vc2402.sdfplugin.psi.SDFType
import com.vc2402.sdfplugin.psi.SDFPsiFile
import com.vc2402.sdfplugin.psi.SDFTypeDeclaration
import com.vc2402.sdfplugin.psi.Types
import org.jetbrains.annotations.NotNull
import org.jsoup.internal.StringUtil
import java.util.*

object Util {
    /**
     * Searches the entire project for Simple language files with instances of the Simple property with the given key.
     *
     * @param project current project
     * @param key     to check
     * @return matching properties
     */
    fun findTypes(project: Project, key: String): List<SDFType> {
        val result: MutableList<SDFType> = ArrayList<SDFType>()
        val virtualFiles: Collection<VirtualFile> =
            FileTypeIndex.getFiles(SDFFileType.INSTANCE, GlobalSearchScope.allScope(project))
        for (virtualFile in virtualFiles) {
            val simpleFile: SDFPsiFile = PsiManager.getInstance(project).findFile(virtualFile) as SDFPsiFile
            val properties: Array<SDFType> = PsiTreeUtil.getChildrenOfType(
                simpleFile,
                SDFType::class.java
            ) as Array<SDFType>
            for (property in properties) {
                if (key == property.text) {
                    result.add(property)
                }
            }
        }
        return result
    }

    fun findTypes(project: Project): List<TypeDescription> {
        val result: MutableList<TypeDescription> = ArrayList<TypeDescription>()
        try {
            val virtualFiles: Collection<VirtualFile> =
                FileTypeIndex.getFiles(SDFFileType.INSTANCE, GlobalSearchScope.allScope(project))
            for (virtualFile in virtualFiles) {
                val f: SDFPsiFile = PsiManager.getInstance(project).findFile(virtualFile) as SDFPsiFile
                result.addAll(f.getTypes())
//                val tree = f.firstChild
//                var pckg = ""
//                var el = tree
//                while(el != null) {
//                    if(el.elementType == Types.PACKAGE_DECLARATION) {
//                        var pe = el.firstChild
//                        while(pe != null && pe.elementType != Types.IDENTIFIER)
//                            pe = pe.nextSibling
//                        if(pe != null)
//                            pckg = pe.text
//                    }
//                    if(el.elementType == Types.DECLARATIONS) {
//                        el = el.firstChild
//                        continue
//                    }
//                    if(el.elementType == Types.DECLARATION &&
//                        el.firstChild != null &&
//                        el.firstChild.elementType == Types.TYPE_DECLARATION) {
//                        lateinit var name: String
//                        var extendable = false
//                        var embeddable = false
//                        var dictionary = false
//                        var de = el.firstChild.firstChild
//                        while(de != null) {
//                            if(de.elementType == Types.TYPE_MODIFIERS) {
//                                var tm = de.firstChild
//                                while(tm != null) {
//                                    if(tm.firstChild != null && tm.firstChild.elementType == Types.TYPEMODIFIER) {
//                                        if(tm.text == "extendable")
//                                            extendable = true
//                                        else if(tm.text == "dictionary")
//                                            dictionary = true
//                                        else if(tm.text == "embeddable")
//                                            embeddable = true
//                                    }
//                                    tm = tm.nextSibling
//                                }
//                            }
//                            if(de.elementType == Types.KW_TYPE) {
//                                var ee = de.nextSibling
//                                while(ee != null) {
//                                    if (ee.elementType == Types.IDENTIFIER) {
//                                        name = ee.text
//                                        break
//                                    }
//                                    ee = ee.nextSibling
//                                }
//                                break
//                            }
//                            de = de.nextSibling
//                        }
//                        result.add(TypeDescription(pckg, name, extendable, embeddable, dictionary))
//                    }
//                    el = el.nextSibling
//                }
            }
        } catch (t: Throwable) {
            val e = t
        }
        return result
    }

    /**
     * Attempts to collect any comment elements above the Simple key/value pair.
     */
    @NotNull
    fun findDocumentationComment(property: SDFType): String {
        val result: MutableList<String> = LinkedList()
        var element: PsiElement = property.prevSibling
        while (element is PsiComment || element is PsiWhiteSpace) {
            if (element is PsiComment) {
                val commentText: String = element.getText()
                result.add(commentText)
            }
            element = element.prevSibling
        }
        return StringUtil.join(Lists.reverse(result), "\n ")
    }
}