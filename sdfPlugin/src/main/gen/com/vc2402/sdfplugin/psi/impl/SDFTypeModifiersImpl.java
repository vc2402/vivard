// This is a generated file. Not intended for manual editing.
package com.vc2402.sdfplugin.psi.impl;

import java.util.List;
import org.jetbrains.annotations.*;
import com.intellij.lang.ASTNode;
import com.intellij.psi.PsiElement;
import com.intellij.psi.PsiElementVisitor;
import com.intellij.psi.util.PsiTreeUtil;
import static com.vc2402.sdfplugin.psi.Types.*;
import com.intellij.extapi.psi.ASTWrapperPsiElement;
import com.vc2402.sdfplugin.psi.*;

public class SDFTypeModifiersImpl extends ASTWrapperPsiElement implements SDFTypeModifiers {

  public SDFTypeModifiersImpl(@NotNull ASTNode node) {
    super(node);
  }

  public void accept(@NotNull SDFVisitor visitor) {
    visitor.visitTypeModifiers(this);
  }

  @Override
  public void accept(@NotNull PsiElementVisitor visitor) {
    if (visitor instanceof SDFVisitor) accept((SDFVisitor)visitor);
    else super.accept(visitor);
  }

  @Override
  @NotNull
  public List<SDFTypeModifier> getTypeModifierList() {
    return PsiTreeUtil.getChildrenOfTypeAsList(this, SDFTypeModifier.class);
  }

}
