// Code generated from Java20Parser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package java20 // Java20Parser
import "github.com/antlr4-go/antlr/v4"

// BaseJava20ParserListener is a complete listener for a parse tree produced by Java20Parser.
type BaseJava20ParserListener struct{}

var _ Java20ParserListener = &BaseJava20ParserListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseJava20ParserListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseJava20ParserListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseJava20ParserListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseJava20ParserListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterStart_ is called when production start_ is entered.
func (s *BaseJava20ParserListener) EnterStart_(ctx *Start_Context) {}

// ExitStart_ is called when production start_ is exited.
func (s *BaseJava20ParserListener) ExitStart_(ctx *Start_Context) {}

// EnterIdentifier is called when production identifier is entered.
func (s *BaseJava20ParserListener) EnterIdentifier(ctx *IdentifierContext) {}

// ExitIdentifier is called when production identifier is exited.
func (s *BaseJava20ParserListener) ExitIdentifier(ctx *IdentifierContext) {}

// EnterTypeIdentifier is called when production typeIdentifier is entered.
func (s *BaseJava20ParserListener) EnterTypeIdentifier(ctx *TypeIdentifierContext) {}

// ExitTypeIdentifier is called when production typeIdentifier is exited.
func (s *BaseJava20ParserListener) ExitTypeIdentifier(ctx *TypeIdentifierContext) {}

// EnterUnqualifiedMethodIdentifier is called when production unqualifiedMethodIdentifier is entered.
func (s *BaseJava20ParserListener) EnterUnqualifiedMethodIdentifier(ctx *UnqualifiedMethodIdentifierContext) {
}

// ExitUnqualifiedMethodIdentifier is called when production unqualifiedMethodIdentifier is exited.
func (s *BaseJava20ParserListener) ExitUnqualifiedMethodIdentifier(ctx *UnqualifiedMethodIdentifierContext) {
}

// EnterContextualKeyword is called when production contextualKeyword is entered.
func (s *BaseJava20ParserListener) EnterContextualKeyword(ctx *ContextualKeywordContext) {}

// ExitContextualKeyword is called when production contextualKeyword is exited.
func (s *BaseJava20ParserListener) ExitContextualKeyword(ctx *ContextualKeywordContext) {}

// EnterContextualKeywordMinusForTypeIdentifier is called when production contextualKeywordMinusForTypeIdentifier is entered.
func (s *BaseJava20ParserListener) EnterContextualKeywordMinusForTypeIdentifier(ctx *ContextualKeywordMinusForTypeIdentifierContext) {
}

// ExitContextualKeywordMinusForTypeIdentifier is called when production contextualKeywordMinusForTypeIdentifier is exited.
func (s *BaseJava20ParserListener) ExitContextualKeywordMinusForTypeIdentifier(ctx *ContextualKeywordMinusForTypeIdentifierContext) {
}

// EnterContextualKeywordMinusForUnqualifiedMethodIdentifier is called when production contextualKeywordMinusForUnqualifiedMethodIdentifier is entered.
func (s *BaseJava20ParserListener) EnterContextualKeywordMinusForUnqualifiedMethodIdentifier(ctx *ContextualKeywordMinusForUnqualifiedMethodIdentifierContext) {
}

// ExitContextualKeywordMinusForUnqualifiedMethodIdentifier is called when production contextualKeywordMinusForUnqualifiedMethodIdentifier is exited.
func (s *BaseJava20ParserListener) ExitContextualKeywordMinusForUnqualifiedMethodIdentifier(ctx *ContextualKeywordMinusForUnqualifiedMethodIdentifierContext) {
}

// EnterLiteral is called when production literal is entered.
func (s *BaseJava20ParserListener) EnterLiteral(ctx *LiteralContext) {}

// ExitLiteral is called when production literal is exited.
func (s *BaseJava20ParserListener) ExitLiteral(ctx *LiteralContext) {}

// EnterPrimitiveType is called when production primitiveType is entered.
func (s *BaseJava20ParserListener) EnterPrimitiveType(ctx *PrimitiveTypeContext) {}

// ExitPrimitiveType is called when production primitiveType is exited.
func (s *BaseJava20ParserListener) ExitPrimitiveType(ctx *PrimitiveTypeContext) {}

// EnterNumericType is called when production numericType is entered.
func (s *BaseJava20ParserListener) EnterNumericType(ctx *NumericTypeContext) {}

// ExitNumericType is called when production numericType is exited.
func (s *BaseJava20ParserListener) ExitNumericType(ctx *NumericTypeContext) {}

// EnterIntegralType is called when production integralType is entered.
func (s *BaseJava20ParserListener) EnterIntegralType(ctx *IntegralTypeContext) {}

// ExitIntegralType is called when production integralType is exited.
func (s *BaseJava20ParserListener) ExitIntegralType(ctx *IntegralTypeContext) {}

// EnterFloatingPointType is called when production floatingPointType is entered.
func (s *BaseJava20ParserListener) EnterFloatingPointType(ctx *FloatingPointTypeContext) {}

// ExitFloatingPointType is called when production floatingPointType is exited.
func (s *BaseJava20ParserListener) ExitFloatingPointType(ctx *FloatingPointTypeContext) {}

// EnterReferenceType is called when production referenceType is entered.
func (s *BaseJava20ParserListener) EnterReferenceType(ctx *ReferenceTypeContext) {}

// ExitReferenceType is called when production referenceType is exited.
func (s *BaseJava20ParserListener) ExitReferenceType(ctx *ReferenceTypeContext) {}

// EnterCoit is called when production coit is entered.
func (s *BaseJava20ParserListener) EnterCoit(ctx *CoitContext) {}

// ExitCoit is called when production coit is exited.
func (s *BaseJava20ParserListener) ExitCoit(ctx *CoitContext) {}

// EnterClassOrInterfaceType is called when production classOrInterfaceType is entered.
func (s *BaseJava20ParserListener) EnterClassOrInterfaceType(ctx *ClassOrInterfaceTypeContext) {}

// ExitClassOrInterfaceType is called when production classOrInterfaceType is exited.
func (s *BaseJava20ParserListener) ExitClassOrInterfaceType(ctx *ClassOrInterfaceTypeContext) {}

// EnterClassType is called when production classType is entered.
func (s *BaseJava20ParserListener) EnterClassType(ctx *ClassTypeContext) {}

// ExitClassType is called when production classType is exited.
func (s *BaseJava20ParserListener) ExitClassType(ctx *ClassTypeContext) {}

// EnterInterfaceType is called when production interfaceType is entered.
func (s *BaseJava20ParserListener) EnterInterfaceType(ctx *InterfaceTypeContext) {}

// ExitInterfaceType is called when production interfaceType is exited.
func (s *BaseJava20ParserListener) ExitInterfaceType(ctx *InterfaceTypeContext) {}

// EnterTypeVariable is called when production typeVariable is entered.
func (s *BaseJava20ParserListener) EnterTypeVariable(ctx *TypeVariableContext) {}

// ExitTypeVariable is called when production typeVariable is exited.
func (s *BaseJava20ParserListener) ExitTypeVariable(ctx *TypeVariableContext) {}

// EnterArrayType is called when production arrayType is entered.
func (s *BaseJava20ParserListener) EnterArrayType(ctx *ArrayTypeContext) {}

// ExitArrayType is called when production arrayType is exited.
func (s *BaseJava20ParserListener) ExitArrayType(ctx *ArrayTypeContext) {}

// EnterDims is called when production dims is entered.
func (s *BaseJava20ParserListener) EnterDims(ctx *DimsContext) {}

// ExitDims is called when production dims is exited.
func (s *BaseJava20ParserListener) ExitDims(ctx *DimsContext) {}

// EnterTypeParameter is called when production typeParameter is entered.
func (s *BaseJava20ParserListener) EnterTypeParameter(ctx *TypeParameterContext) {}

// ExitTypeParameter is called when production typeParameter is exited.
func (s *BaseJava20ParserListener) ExitTypeParameter(ctx *TypeParameterContext) {}

// EnterTypeParameterModifier is called when production typeParameterModifier is entered.
func (s *BaseJava20ParserListener) EnterTypeParameterModifier(ctx *TypeParameterModifierContext) {}

// ExitTypeParameterModifier is called when production typeParameterModifier is exited.
func (s *BaseJava20ParserListener) ExitTypeParameterModifier(ctx *TypeParameterModifierContext) {}

// EnterTypeBound is called when production typeBound is entered.
func (s *BaseJava20ParserListener) EnterTypeBound(ctx *TypeBoundContext) {}

// ExitTypeBound is called when production typeBound is exited.
func (s *BaseJava20ParserListener) ExitTypeBound(ctx *TypeBoundContext) {}

// EnterAdditionalBound is called when production additionalBound is entered.
func (s *BaseJava20ParserListener) EnterAdditionalBound(ctx *AdditionalBoundContext) {}

// ExitAdditionalBound is called when production additionalBound is exited.
func (s *BaseJava20ParserListener) ExitAdditionalBound(ctx *AdditionalBoundContext) {}

// EnterTypeArguments is called when production typeArguments is entered.
func (s *BaseJava20ParserListener) EnterTypeArguments(ctx *TypeArgumentsContext) {}

// ExitTypeArguments is called when production typeArguments is exited.
func (s *BaseJava20ParserListener) ExitTypeArguments(ctx *TypeArgumentsContext) {}

// EnterTypeArgumentList is called when production typeArgumentList is entered.
func (s *BaseJava20ParserListener) EnterTypeArgumentList(ctx *TypeArgumentListContext) {}

// ExitTypeArgumentList is called when production typeArgumentList is exited.
func (s *BaseJava20ParserListener) ExitTypeArgumentList(ctx *TypeArgumentListContext) {}

// EnterTypeArgument is called when production typeArgument is entered.
func (s *BaseJava20ParserListener) EnterTypeArgument(ctx *TypeArgumentContext) {}

// ExitTypeArgument is called when production typeArgument is exited.
func (s *BaseJava20ParserListener) ExitTypeArgument(ctx *TypeArgumentContext) {}

// EnterWildcard is called when production wildcard is entered.
func (s *BaseJava20ParserListener) EnterWildcard(ctx *WildcardContext) {}

// ExitWildcard is called when production wildcard is exited.
func (s *BaseJava20ParserListener) ExitWildcard(ctx *WildcardContext) {}

// EnterWildcardBounds is called when production wildcardBounds is entered.
func (s *BaseJava20ParserListener) EnterWildcardBounds(ctx *WildcardBoundsContext) {}

// ExitWildcardBounds is called when production wildcardBounds is exited.
func (s *BaseJava20ParserListener) ExitWildcardBounds(ctx *WildcardBoundsContext) {}

// EnterModuleName is called when production moduleName is entered.
func (s *BaseJava20ParserListener) EnterModuleName(ctx *ModuleNameContext) {}

// ExitModuleName is called when production moduleName is exited.
func (s *BaseJava20ParserListener) ExitModuleName(ctx *ModuleNameContext) {}

// EnterPackageName is called when production packageName is entered.
func (s *BaseJava20ParserListener) EnterPackageName(ctx *PackageNameContext) {}

// ExitPackageName is called when production packageName is exited.
func (s *BaseJava20ParserListener) ExitPackageName(ctx *PackageNameContext) {}

// EnterTypeName is called when production typeName is entered.
func (s *BaseJava20ParserListener) EnterTypeName(ctx *TypeNameContext) {}

// ExitTypeName is called when production typeName is exited.
func (s *BaseJava20ParserListener) ExitTypeName(ctx *TypeNameContext) {}

// EnterPackageOrTypeName is called when production packageOrTypeName is entered.
func (s *BaseJava20ParserListener) EnterPackageOrTypeName(ctx *PackageOrTypeNameContext) {}

// ExitPackageOrTypeName is called when production packageOrTypeName is exited.
func (s *BaseJava20ParserListener) ExitPackageOrTypeName(ctx *PackageOrTypeNameContext) {}

// EnterExpressionName is called when production expressionName is entered.
func (s *BaseJava20ParserListener) EnterExpressionName(ctx *ExpressionNameContext) {}

// ExitExpressionName is called when production expressionName is exited.
func (s *BaseJava20ParserListener) ExitExpressionName(ctx *ExpressionNameContext) {}

// EnterMethodName is called when production methodName is entered.
func (s *BaseJava20ParserListener) EnterMethodName(ctx *MethodNameContext) {}

// ExitMethodName is called when production methodName is exited.
func (s *BaseJava20ParserListener) ExitMethodName(ctx *MethodNameContext) {}

// EnterAmbiguousName is called when production ambiguousName is entered.
func (s *BaseJava20ParserListener) EnterAmbiguousName(ctx *AmbiguousNameContext) {}

// ExitAmbiguousName is called when production ambiguousName is exited.
func (s *BaseJava20ParserListener) ExitAmbiguousName(ctx *AmbiguousNameContext) {}

// EnterCompilationUnit is called when production compilationUnit is entered.
func (s *BaseJava20ParserListener) EnterCompilationUnit(ctx *CompilationUnitContext) {}

// ExitCompilationUnit is called when production compilationUnit is exited.
func (s *BaseJava20ParserListener) ExitCompilationUnit(ctx *CompilationUnitContext) {}

// EnterOrdinaryCompilationUnit is called when production ordinaryCompilationUnit is entered.
func (s *BaseJava20ParserListener) EnterOrdinaryCompilationUnit(ctx *OrdinaryCompilationUnitContext) {
}

// ExitOrdinaryCompilationUnit is called when production ordinaryCompilationUnit is exited.
func (s *BaseJava20ParserListener) ExitOrdinaryCompilationUnit(ctx *OrdinaryCompilationUnitContext) {}

// EnterModularCompilationUnit is called when production modularCompilationUnit is entered.
func (s *BaseJava20ParserListener) EnterModularCompilationUnit(ctx *ModularCompilationUnitContext) {}

// ExitModularCompilationUnit is called when production modularCompilationUnit is exited.
func (s *BaseJava20ParserListener) ExitModularCompilationUnit(ctx *ModularCompilationUnitContext) {}

// EnterPackageDeclaration is called when production packageDeclaration is entered.
func (s *BaseJava20ParserListener) EnterPackageDeclaration(ctx *PackageDeclarationContext) {}

// ExitPackageDeclaration is called when production packageDeclaration is exited.
func (s *BaseJava20ParserListener) ExitPackageDeclaration(ctx *PackageDeclarationContext) {}

// EnterPackageModifier is called when production packageModifier is entered.
func (s *BaseJava20ParserListener) EnterPackageModifier(ctx *PackageModifierContext) {}

// ExitPackageModifier is called when production packageModifier is exited.
func (s *BaseJava20ParserListener) ExitPackageModifier(ctx *PackageModifierContext) {}

// EnterImportDeclaration is called when production importDeclaration is entered.
func (s *BaseJava20ParserListener) EnterImportDeclaration(ctx *ImportDeclarationContext) {}

// ExitImportDeclaration is called when production importDeclaration is exited.
func (s *BaseJava20ParserListener) ExitImportDeclaration(ctx *ImportDeclarationContext) {}

// EnterSingleTypeImportDeclaration is called when production singleTypeImportDeclaration is entered.
func (s *BaseJava20ParserListener) EnterSingleTypeImportDeclaration(ctx *SingleTypeImportDeclarationContext) {
}

// ExitSingleTypeImportDeclaration is called when production singleTypeImportDeclaration is exited.
func (s *BaseJava20ParserListener) ExitSingleTypeImportDeclaration(ctx *SingleTypeImportDeclarationContext) {
}

// EnterTypeImportOnDemandDeclaration is called when production typeImportOnDemandDeclaration is entered.
func (s *BaseJava20ParserListener) EnterTypeImportOnDemandDeclaration(ctx *TypeImportOnDemandDeclarationContext) {
}

// ExitTypeImportOnDemandDeclaration is called when production typeImportOnDemandDeclaration is exited.
func (s *BaseJava20ParserListener) ExitTypeImportOnDemandDeclaration(ctx *TypeImportOnDemandDeclarationContext) {
}

// EnterSingleStaticImportDeclaration is called when production singleStaticImportDeclaration is entered.
func (s *BaseJava20ParserListener) EnterSingleStaticImportDeclaration(ctx *SingleStaticImportDeclarationContext) {
}

// ExitSingleStaticImportDeclaration is called when production singleStaticImportDeclaration is exited.
func (s *BaseJava20ParserListener) ExitSingleStaticImportDeclaration(ctx *SingleStaticImportDeclarationContext) {
}

// EnterStaticImportOnDemandDeclaration is called when production staticImportOnDemandDeclaration is entered.
func (s *BaseJava20ParserListener) EnterStaticImportOnDemandDeclaration(ctx *StaticImportOnDemandDeclarationContext) {
}

// ExitStaticImportOnDemandDeclaration is called when production staticImportOnDemandDeclaration is exited.
func (s *BaseJava20ParserListener) ExitStaticImportOnDemandDeclaration(ctx *StaticImportOnDemandDeclarationContext) {
}

// EnterTopLevelClassOrInterfaceDeclaration is called when production topLevelClassOrInterfaceDeclaration is entered.
func (s *BaseJava20ParserListener) EnterTopLevelClassOrInterfaceDeclaration(ctx *TopLevelClassOrInterfaceDeclarationContext) {
}

// ExitTopLevelClassOrInterfaceDeclaration is called when production topLevelClassOrInterfaceDeclaration is exited.
func (s *BaseJava20ParserListener) ExitTopLevelClassOrInterfaceDeclaration(ctx *TopLevelClassOrInterfaceDeclarationContext) {
}

// EnterModuleDeclaration is called when production moduleDeclaration is entered.
func (s *BaseJava20ParserListener) EnterModuleDeclaration(ctx *ModuleDeclarationContext) {}

// ExitModuleDeclaration is called when production moduleDeclaration is exited.
func (s *BaseJava20ParserListener) ExitModuleDeclaration(ctx *ModuleDeclarationContext) {}

// EnterModuleDirective is called when production moduleDirective is entered.
func (s *BaseJava20ParserListener) EnterModuleDirective(ctx *ModuleDirectiveContext) {}

// ExitModuleDirective is called when production moduleDirective is exited.
func (s *BaseJava20ParserListener) ExitModuleDirective(ctx *ModuleDirectiveContext) {}

// EnterRequiresModifier is called when production requiresModifier is entered.
func (s *BaseJava20ParserListener) EnterRequiresModifier(ctx *RequiresModifierContext) {}

// ExitRequiresModifier is called when production requiresModifier is exited.
func (s *BaseJava20ParserListener) ExitRequiresModifier(ctx *RequiresModifierContext) {}

// EnterClassDeclaration is called when production classDeclaration is entered.
func (s *BaseJava20ParserListener) EnterClassDeclaration(ctx *ClassDeclarationContext) {}

// ExitClassDeclaration is called when production classDeclaration is exited.
func (s *BaseJava20ParserListener) ExitClassDeclaration(ctx *ClassDeclarationContext) {}

// EnterNormalClassDeclaration is called when production normalClassDeclaration is entered.
func (s *BaseJava20ParserListener) EnterNormalClassDeclaration(ctx *NormalClassDeclarationContext) {}

// ExitNormalClassDeclaration is called when production normalClassDeclaration is exited.
func (s *BaseJava20ParserListener) ExitNormalClassDeclaration(ctx *NormalClassDeclarationContext) {}

// EnterClassModifier is called when production classModifier is entered.
func (s *BaseJava20ParserListener) EnterClassModifier(ctx *ClassModifierContext) {}

// ExitClassModifier is called when production classModifier is exited.
func (s *BaseJava20ParserListener) ExitClassModifier(ctx *ClassModifierContext) {}

// EnterTypeParameters is called when production typeParameters is entered.
func (s *BaseJava20ParserListener) EnterTypeParameters(ctx *TypeParametersContext) {}

// ExitTypeParameters is called when production typeParameters is exited.
func (s *BaseJava20ParserListener) ExitTypeParameters(ctx *TypeParametersContext) {}

// EnterTypeParameterList is called when production typeParameterList is entered.
func (s *BaseJava20ParserListener) EnterTypeParameterList(ctx *TypeParameterListContext) {}

// ExitTypeParameterList is called when production typeParameterList is exited.
func (s *BaseJava20ParserListener) ExitTypeParameterList(ctx *TypeParameterListContext) {}

// EnterClassExtends is called when production classExtends is entered.
func (s *BaseJava20ParserListener) EnterClassExtends(ctx *ClassExtendsContext) {}

// ExitClassExtends is called when production classExtends is exited.
func (s *BaseJava20ParserListener) ExitClassExtends(ctx *ClassExtendsContext) {}

// EnterClassImplements is called when production classImplements is entered.
func (s *BaseJava20ParserListener) EnterClassImplements(ctx *ClassImplementsContext) {}

// ExitClassImplements is called when production classImplements is exited.
func (s *BaseJava20ParserListener) ExitClassImplements(ctx *ClassImplementsContext) {}

// EnterInterfaceTypeList is called when production interfaceTypeList is entered.
func (s *BaseJava20ParserListener) EnterInterfaceTypeList(ctx *InterfaceTypeListContext) {}

// ExitInterfaceTypeList is called when production interfaceTypeList is exited.
func (s *BaseJava20ParserListener) ExitInterfaceTypeList(ctx *InterfaceTypeListContext) {}

// EnterClassPermits is called when production classPermits is entered.
func (s *BaseJava20ParserListener) EnterClassPermits(ctx *ClassPermitsContext) {}

// ExitClassPermits is called when production classPermits is exited.
func (s *BaseJava20ParserListener) ExitClassPermits(ctx *ClassPermitsContext) {}

// EnterClassBody is called when production classBody is entered.
func (s *BaseJava20ParserListener) EnterClassBody(ctx *ClassBodyContext) {}

// ExitClassBody is called when production classBody is exited.
func (s *BaseJava20ParserListener) ExitClassBody(ctx *ClassBodyContext) {}

// EnterClassBodyDeclaration is called when production classBodyDeclaration is entered.
func (s *BaseJava20ParserListener) EnterClassBodyDeclaration(ctx *ClassBodyDeclarationContext) {}

// ExitClassBodyDeclaration is called when production classBodyDeclaration is exited.
func (s *BaseJava20ParserListener) ExitClassBodyDeclaration(ctx *ClassBodyDeclarationContext) {}

// EnterClassMemberDeclaration is called when production classMemberDeclaration is entered.
func (s *BaseJava20ParserListener) EnterClassMemberDeclaration(ctx *ClassMemberDeclarationContext) {}

// ExitClassMemberDeclaration is called when production classMemberDeclaration is exited.
func (s *BaseJava20ParserListener) ExitClassMemberDeclaration(ctx *ClassMemberDeclarationContext) {}

// EnterFieldDeclaration is called when production fieldDeclaration is entered.
func (s *BaseJava20ParserListener) EnterFieldDeclaration(ctx *FieldDeclarationContext) {}

// ExitFieldDeclaration is called when production fieldDeclaration is exited.
func (s *BaseJava20ParserListener) ExitFieldDeclaration(ctx *FieldDeclarationContext) {}

// EnterFieldModifier is called when production fieldModifier is entered.
func (s *BaseJava20ParserListener) EnterFieldModifier(ctx *FieldModifierContext) {}

// ExitFieldModifier is called when production fieldModifier is exited.
func (s *BaseJava20ParserListener) ExitFieldModifier(ctx *FieldModifierContext) {}

// EnterVariableDeclaratorList is called when production variableDeclaratorList is entered.
func (s *BaseJava20ParserListener) EnterVariableDeclaratorList(ctx *VariableDeclaratorListContext) {}

// ExitVariableDeclaratorList is called when production variableDeclaratorList is exited.
func (s *BaseJava20ParserListener) ExitVariableDeclaratorList(ctx *VariableDeclaratorListContext) {}

// EnterVariableDeclarator is called when production variableDeclarator is entered.
func (s *BaseJava20ParserListener) EnterVariableDeclarator(ctx *VariableDeclaratorContext) {}

// ExitVariableDeclarator is called when production variableDeclarator is exited.
func (s *BaseJava20ParserListener) ExitVariableDeclarator(ctx *VariableDeclaratorContext) {}

// EnterVariableDeclaratorId is called when production variableDeclaratorId is entered.
func (s *BaseJava20ParserListener) EnterVariableDeclaratorId(ctx *VariableDeclaratorIdContext) {}

// ExitVariableDeclaratorId is called when production variableDeclaratorId is exited.
func (s *BaseJava20ParserListener) ExitVariableDeclaratorId(ctx *VariableDeclaratorIdContext) {}

// EnterVariableInitializer is called when production variableInitializer is entered.
func (s *BaseJava20ParserListener) EnterVariableInitializer(ctx *VariableInitializerContext) {}

// ExitVariableInitializer is called when production variableInitializer is exited.
func (s *BaseJava20ParserListener) ExitVariableInitializer(ctx *VariableInitializerContext) {}

// EnterUnannType is called when production unannType is entered.
func (s *BaseJava20ParserListener) EnterUnannType(ctx *UnannTypeContext) {}

// ExitUnannType is called when production unannType is exited.
func (s *BaseJava20ParserListener) ExitUnannType(ctx *UnannTypeContext) {}

// EnterUnannPrimitiveType is called when production unannPrimitiveType is entered.
func (s *BaseJava20ParserListener) EnterUnannPrimitiveType(ctx *UnannPrimitiveTypeContext) {}

// ExitUnannPrimitiveType is called when production unannPrimitiveType is exited.
func (s *BaseJava20ParserListener) ExitUnannPrimitiveType(ctx *UnannPrimitiveTypeContext) {}

// EnterUnannReferenceType is called when production unannReferenceType is entered.
func (s *BaseJava20ParserListener) EnterUnannReferenceType(ctx *UnannReferenceTypeContext) {}

// ExitUnannReferenceType is called when production unannReferenceType is exited.
func (s *BaseJava20ParserListener) ExitUnannReferenceType(ctx *UnannReferenceTypeContext) {}

// EnterUnannClassOrInterfaceType is called when production unannClassOrInterfaceType is entered.
func (s *BaseJava20ParserListener) EnterUnannClassOrInterfaceType(ctx *UnannClassOrInterfaceTypeContext) {
}

// ExitUnannClassOrInterfaceType is called when production unannClassOrInterfaceType is exited.
func (s *BaseJava20ParserListener) ExitUnannClassOrInterfaceType(ctx *UnannClassOrInterfaceTypeContext) {
}

// EnterUCOIT is called when production uCOIT is entered.
func (s *BaseJava20ParserListener) EnterUCOIT(ctx *UCOITContext) {}

// ExitUCOIT is called when production uCOIT is exited.
func (s *BaseJava20ParserListener) ExitUCOIT(ctx *UCOITContext) {}

// EnterUnannClassType is called when production unannClassType is entered.
func (s *BaseJava20ParserListener) EnterUnannClassType(ctx *UnannClassTypeContext) {}

// ExitUnannClassType is called when production unannClassType is exited.
func (s *BaseJava20ParserListener) ExitUnannClassType(ctx *UnannClassTypeContext) {}

// EnterUnannInterfaceType is called when production unannInterfaceType is entered.
func (s *BaseJava20ParserListener) EnterUnannInterfaceType(ctx *UnannInterfaceTypeContext) {}

// ExitUnannInterfaceType is called when production unannInterfaceType is exited.
func (s *BaseJava20ParserListener) ExitUnannInterfaceType(ctx *UnannInterfaceTypeContext) {}

// EnterUnannTypeVariable is called when production unannTypeVariable is entered.
func (s *BaseJava20ParserListener) EnterUnannTypeVariable(ctx *UnannTypeVariableContext) {}

// ExitUnannTypeVariable is called when production unannTypeVariable is exited.
func (s *BaseJava20ParserListener) ExitUnannTypeVariable(ctx *UnannTypeVariableContext) {}

// EnterUnannArrayType is called when production unannArrayType is entered.
func (s *BaseJava20ParserListener) EnterUnannArrayType(ctx *UnannArrayTypeContext) {}

// ExitUnannArrayType is called when production unannArrayType is exited.
func (s *BaseJava20ParserListener) ExitUnannArrayType(ctx *UnannArrayTypeContext) {}

// EnterMethodDeclaration is called when production methodDeclaration is entered.
func (s *BaseJava20ParserListener) EnterMethodDeclaration(ctx *MethodDeclarationContext) {}

// ExitMethodDeclaration is called when production methodDeclaration is exited.
func (s *BaseJava20ParserListener) ExitMethodDeclaration(ctx *MethodDeclarationContext) {}

// EnterMethodModifier is called when production methodModifier is entered.
func (s *BaseJava20ParserListener) EnterMethodModifier(ctx *MethodModifierContext) {}

// ExitMethodModifier is called when production methodModifier is exited.
func (s *BaseJava20ParserListener) ExitMethodModifier(ctx *MethodModifierContext) {}

// EnterMethodHeader is called when production methodHeader is entered.
func (s *BaseJava20ParserListener) EnterMethodHeader(ctx *MethodHeaderContext) {}

// ExitMethodHeader is called when production methodHeader is exited.
func (s *BaseJava20ParserListener) ExitMethodHeader(ctx *MethodHeaderContext) {}

// EnterResult is called when production result is entered.
func (s *BaseJava20ParserListener) EnterResult(ctx *ResultContext) {}

// ExitResult is called when production result is exited.
func (s *BaseJava20ParserListener) ExitResult(ctx *ResultContext) {}

// EnterMethodDeclarator is called when production methodDeclarator is entered.
func (s *BaseJava20ParserListener) EnterMethodDeclarator(ctx *MethodDeclaratorContext) {}

// ExitMethodDeclarator is called when production methodDeclarator is exited.
func (s *BaseJava20ParserListener) ExitMethodDeclarator(ctx *MethodDeclaratorContext) {}

// EnterReceiverParameter is called when production receiverParameter is entered.
func (s *BaseJava20ParserListener) EnterReceiverParameter(ctx *ReceiverParameterContext) {}

// ExitReceiverParameter is called when production receiverParameter is exited.
func (s *BaseJava20ParserListener) ExitReceiverParameter(ctx *ReceiverParameterContext) {}

// EnterFormalParameterList is called when production formalParameterList is entered.
func (s *BaseJava20ParserListener) EnterFormalParameterList(ctx *FormalParameterListContext) {}

// ExitFormalParameterList is called when production formalParameterList is exited.
func (s *BaseJava20ParserListener) ExitFormalParameterList(ctx *FormalParameterListContext) {}

// EnterFormalParameter is called when production formalParameter is entered.
func (s *BaseJava20ParserListener) EnterFormalParameter(ctx *FormalParameterContext) {}

// ExitFormalParameter is called when production formalParameter is exited.
func (s *BaseJava20ParserListener) ExitFormalParameter(ctx *FormalParameterContext) {}

// EnterVariableArityParameter is called when production variableArityParameter is entered.
func (s *BaseJava20ParserListener) EnterVariableArityParameter(ctx *VariableArityParameterContext) {}

// ExitVariableArityParameter is called when production variableArityParameter is exited.
func (s *BaseJava20ParserListener) ExitVariableArityParameter(ctx *VariableArityParameterContext) {}

// EnterVariableModifier is called when production variableModifier is entered.
func (s *BaseJava20ParserListener) EnterVariableModifier(ctx *VariableModifierContext) {}

// ExitVariableModifier is called when production variableModifier is exited.
func (s *BaseJava20ParserListener) ExitVariableModifier(ctx *VariableModifierContext) {}

// EnterThrowsT is called when production throwsT is entered.
func (s *BaseJava20ParserListener) EnterThrowsT(ctx *ThrowsTContext) {}

// ExitThrowsT is called when production throwsT is exited.
func (s *BaseJava20ParserListener) ExitThrowsT(ctx *ThrowsTContext) {}

// EnterExceptionTypeList is called when production exceptionTypeList is entered.
func (s *BaseJava20ParserListener) EnterExceptionTypeList(ctx *ExceptionTypeListContext) {}

// ExitExceptionTypeList is called when production exceptionTypeList is exited.
func (s *BaseJava20ParserListener) ExitExceptionTypeList(ctx *ExceptionTypeListContext) {}

// EnterExceptionType is called when production exceptionType is entered.
func (s *BaseJava20ParserListener) EnterExceptionType(ctx *ExceptionTypeContext) {}

// ExitExceptionType is called when production exceptionType is exited.
func (s *BaseJava20ParserListener) ExitExceptionType(ctx *ExceptionTypeContext) {}

// EnterMethodBody is called when production methodBody is entered.
func (s *BaseJava20ParserListener) EnterMethodBody(ctx *MethodBodyContext) {}

// ExitMethodBody is called when production methodBody is exited.
func (s *BaseJava20ParserListener) ExitMethodBody(ctx *MethodBodyContext) {}

// EnterInstanceInitializer is called when production instanceInitializer is entered.
func (s *BaseJava20ParserListener) EnterInstanceInitializer(ctx *InstanceInitializerContext) {}

// ExitInstanceInitializer is called when production instanceInitializer is exited.
func (s *BaseJava20ParserListener) ExitInstanceInitializer(ctx *InstanceInitializerContext) {}

// EnterStaticInitializer is called when production staticInitializer is entered.
func (s *BaseJava20ParserListener) EnterStaticInitializer(ctx *StaticInitializerContext) {}

// ExitStaticInitializer is called when production staticInitializer is exited.
func (s *BaseJava20ParserListener) ExitStaticInitializer(ctx *StaticInitializerContext) {}

// EnterConstructorDeclaration is called when production constructorDeclaration is entered.
func (s *BaseJava20ParserListener) EnterConstructorDeclaration(ctx *ConstructorDeclarationContext) {}

// ExitConstructorDeclaration is called when production constructorDeclaration is exited.
func (s *BaseJava20ParserListener) ExitConstructorDeclaration(ctx *ConstructorDeclarationContext) {}

// EnterConstructorModifier is called when production constructorModifier is entered.
func (s *BaseJava20ParserListener) EnterConstructorModifier(ctx *ConstructorModifierContext) {}

// ExitConstructorModifier is called when production constructorModifier is exited.
func (s *BaseJava20ParserListener) ExitConstructorModifier(ctx *ConstructorModifierContext) {}

// EnterConstructorDeclarator is called when production constructorDeclarator is entered.
func (s *BaseJava20ParserListener) EnterConstructorDeclarator(ctx *ConstructorDeclaratorContext) {}

// ExitConstructorDeclarator is called when production constructorDeclarator is exited.
func (s *BaseJava20ParserListener) ExitConstructorDeclarator(ctx *ConstructorDeclaratorContext) {}

// EnterSimpleTypeName is called when production simpleTypeName is entered.
func (s *BaseJava20ParserListener) EnterSimpleTypeName(ctx *SimpleTypeNameContext) {}

// ExitSimpleTypeName is called when production simpleTypeName is exited.
func (s *BaseJava20ParserListener) ExitSimpleTypeName(ctx *SimpleTypeNameContext) {}

// EnterConstructorBody is called when production constructorBody is entered.
func (s *BaseJava20ParserListener) EnterConstructorBody(ctx *ConstructorBodyContext) {}

// ExitConstructorBody is called when production constructorBody is exited.
func (s *BaseJava20ParserListener) ExitConstructorBody(ctx *ConstructorBodyContext) {}

// EnterExplicitConstructorInvocation is called when production explicitConstructorInvocation is entered.
func (s *BaseJava20ParserListener) EnterExplicitConstructorInvocation(ctx *ExplicitConstructorInvocationContext) {
}

// ExitExplicitConstructorInvocation is called when production explicitConstructorInvocation is exited.
func (s *BaseJava20ParserListener) ExitExplicitConstructorInvocation(ctx *ExplicitConstructorInvocationContext) {
}

// EnterEnumDeclaration is called when production enumDeclaration is entered.
func (s *BaseJava20ParserListener) EnterEnumDeclaration(ctx *EnumDeclarationContext) {}

// ExitEnumDeclaration is called when production enumDeclaration is exited.
func (s *BaseJava20ParserListener) ExitEnumDeclaration(ctx *EnumDeclarationContext) {}

// EnterEnumBody is called when production enumBody is entered.
func (s *BaseJava20ParserListener) EnterEnumBody(ctx *EnumBodyContext) {}

// ExitEnumBody is called when production enumBody is exited.
func (s *BaseJava20ParserListener) ExitEnumBody(ctx *EnumBodyContext) {}

// EnterEnumConstantList is called when production enumConstantList is entered.
func (s *BaseJava20ParserListener) EnterEnumConstantList(ctx *EnumConstantListContext) {}

// ExitEnumConstantList is called when production enumConstantList is exited.
func (s *BaseJava20ParserListener) ExitEnumConstantList(ctx *EnumConstantListContext) {}

// EnterEnumConstant is called when production enumConstant is entered.
func (s *BaseJava20ParserListener) EnterEnumConstant(ctx *EnumConstantContext) {}

// ExitEnumConstant is called when production enumConstant is exited.
func (s *BaseJava20ParserListener) ExitEnumConstant(ctx *EnumConstantContext) {}

// EnterEnumConstantModifier is called when production enumConstantModifier is entered.
func (s *BaseJava20ParserListener) EnterEnumConstantModifier(ctx *EnumConstantModifierContext) {}

// ExitEnumConstantModifier is called when production enumConstantModifier is exited.
func (s *BaseJava20ParserListener) ExitEnumConstantModifier(ctx *EnumConstantModifierContext) {}

// EnterEnumBodyDeclarations is called when production enumBodyDeclarations is entered.
func (s *BaseJava20ParserListener) EnterEnumBodyDeclarations(ctx *EnumBodyDeclarationsContext) {}

// ExitEnumBodyDeclarations is called when production enumBodyDeclarations is exited.
func (s *BaseJava20ParserListener) ExitEnumBodyDeclarations(ctx *EnumBodyDeclarationsContext) {}

// EnterRecordDeclaration is called when production recordDeclaration is entered.
func (s *BaseJava20ParserListener) EnterRecordDeclaration(ctx *RecordDeclarationContext) {}

// ExitRecordDeclaration is called when production recordDeclaration is exited.
func (s *BaseJava20ParserListener) ExitRecordDeclaration(ctx *RecordDeclarationContext) {}

// EnterRecordHeader is called when production recordHeader is entered.
func (s *BaseJava20ParserListener) EnterRecordHeader(ctx *RecordHeaderContext) {}

// ExitRecordHeader is called when production recordHeader is exited.
func (s *BaseJava20ParserListener) ExitRecordHeader(ctx *RecordHeaderContext) {}

// EnterRecordComponentList is called when production recordComponentList is entered.
func (s *BaseJava20ParserListener) EnterRecordComponentList(ctx *RecordComponentListContext) {}

// ExitRecordComponentList is called when production recordComponentList is exited.
func (s *BaseJava20ParserListener) ExitRecordComponentList(ctx *RecordComponentListContext) {}

// EnterRecordComponent is called when production recordComponent is entered.
func (s *BaseJava20ParserListener) EnterRecordComponent(ctx *RecordComponentContext) {}

// ExitRecordComponent is called when production recordComponent is exited.
func (s *BaseJava20ParserListener) ExitRecordComponent(ctx *RecordComponentContext) {}

// EnterVariableArityRecordComponent is called when production variableArityRecordComponent is entered.
func (s *BaseJava20ParserListener) EnterVariableArityRecordComponent(ctx *VariableArityRecordComponentContext) {
}

// ExitVariableArityRecordComponent is called when production variableArityRecordComponent is exited.
func (s *BaseJava20ParserListener) ExitVariableArityRecordComponent(ctx *VariableArityRecordComponentContext) {
}

// EnterRecordComponentModifier is called when production recordComponentModifier is entered.
func (s *BaseJava20ParserListener) EnterRecordComponentModifier(ctx *RecordComponentModifierContext) {
}

// ExitRecordComponentModifier is called when production recordComponentModifier is exited.
func (s *BaseJava20ParserListener) ExitRecordComponentModifier(ctx *RecordComponentModifierContext) {}

// EnterRecordBody is called when production recordBody is entered.
func (s *BaseJava20ParserListener) EnterRecordBody(ctx *RecordBodyContext) {}

// ExitRecordBody is called when production recordBody is exited.
func (s *BaseJava20ParserListener) ExitRecordBody(ctx *RecordBodyContext) {}

// EnterRecordBodyDeclaration is called when production recordBodyDeclaration is entered.
func (s *BaseJava20ParserListener) EnterRecordBodyDeclaration(ctx *RecordBodyDeclarationContext) {}

// ExitRecordBodyDeclaration is called when production recordBodyDeclaration is exited.
func (s *BaseJava20ParserListener) ExitRecordBodyDeclaration(ctx *RecordBodyDeclarationContext) {}

// EnterCompactConstructorDeclaration is called when production compactConstructorDeclaration is entered.
func (s *BaseJava20ParserListener) EnterCompactConstructorDeclaration(ctx *CompactConstructorDeclarationContext) {
}

// ExitCompactConstructorDeclaration is called when production compactConstructorDeclaration is exited.
func (s *BaseJava20ParserListener) ExitCompactConstructorDeclaration(ctx *CompactConstructorDeclarationContext) {
}

// EnterInterfaceDeclaration is called when production interfaceDeclaration is entered.
func (s *BaseJava20ParserListener) EnterInterfaceDeclaration(ctx *InterfaceDeclarationContext) {}

// ExitInterfaceDeclaration is called when production interfaceDeclaration is exited.
func (s *BaseJava20ParserListener) ExitInterfaceDeclaration(ctx *InterfaceDeclarationContext) {}

// EnterNormalInterfaceDeclaration is called when production normalInterfaceDeclaration is entered.
func (s *BaseJava20ParserListener) EnterNormalInterfaceDeclaration(ctx *NormalInterfaceDeclarationContext) {
}

// ExitNormalInterfaceDeclaration is called when production normalInterfaceDeclaration is exited.
func (s *BaseJava20ParserListener) ExitNormalInterfaceDeclaration(ctx *NormalInterfaceDeclarationContext) {
}

// EnterInterfaceModifier is called when production interfaceModifier is entered.
func (s *BaseJava20ParserListener) EnterInterfaceModifier(ctx *InterfaceModifierContext) {}

// ExitInterfaceModifier is called when production interfaceModifier is exited.
func (s *BaseJava20ParserListener) ExitInterfaceModifier(ctx *InterfaceModifierContext) {}

// EnterInterfaceExtends is called when production interfaceExtends is entered.
func (s *BaseJava20ParserListener) EnterInterfaceExtends(ctx *InterfaceExtendsContext) {}

// ExitInterfaceExtends is called when production interfaceExtends is exited.
func (s *BaseJava20ParserListener) ExitInterfaceExtends(ctx *InterfaceExtendsContext) {}

// EnterInterfacePermits is called when production interfacePermits is entered.
func (s *BaseJava20ParserListener) EnterInterfacePermits(ctx *InterfacePermitsContext) {}

// ExitInterfacePermits is called when production interfacePermits is exited.
func (s *BaseJava20ParserListener) ExitInterfacePermits(ctx *InterfacePermitsContext) {}

// EnterInterfaceBody is called when production interfaceBody is entered.
func (s *BaseJava20ParserListener) EnterInterfaceBody(ctx *InterfaceBodyContext) {}

// ExitInterfaceBody is called when production interfaceBody is exited.
func (s *BaseJava20ParserListener) ExitInterfaceBody(ctx *InterfaceBodyContext) {}

// EnterInterfaceMemberDeclaration is called when production interfaceMemberDeclaration is entered.
func (s *BaseJava20ParserListener) EnterInterfaceMemberDeclaration(ctx *InterfaceMemberDeclarationContext) {
}

// ExitInterfaceMemberDeclaration is called when production interfaceMemberDeclaration is exited.
func (s *BaseJava20ParserListener) ExitInterfaceMemberDeclaration(ctx *InterfaceMemberDeclarationContext) {
}

// EnterConstantDeclaration is called when production constantDeclaration is entered.
func (s *BaseJava20ParserListener) EnterConstantDeclaration(ctx *ConstantDeclarationContext) {}

// ExitConstantDeclaration is called when production constantDeclaration is exited.
func (s *BaseJava20ParserListener) ExitConstantDeclaration(ctx *ConstantDeclarationContext) {}

// EnterConstantModifier is called when production constantModifier is entered.
func (s *BaseJava20ParserListener) EnterConstantModifier(ctx *ConstantModifierContext) {}

// ExitConstantModifier is called when production constantModifier is exited.
func (s *BaseJava20ParserListener) ExitConstantModifier(ctx *ConstantModifierContext) {}

// EnterInterfaceMethodDeclaration is called when production interfaceMethodDeclaration is entered.
func (s *BaseJava20ParserListener) EnterInterfaceMethodDeclaration(ctx *InterfaceMethodDeclarationContext) {
}

// ExitInterfaceMethodDeclaration is called when production interfaceMethodDeclaration is exited.
func (s *BaseJava20ParserListener) ExitInterfaceMethodDeclaration(ctx *InterfaceMethodDeclarationContext) {
}

// EnterInterfaceMethodModifier is called when production interfaceMethodModifier is entered.
func (s *BaseJava20ParserListener) EnterInterfaceMethodModifier(ctx *InterfaceMethodModifierContext) {
}

// ExitInterfaceMethodModifier is called when production interfaceMethodModifier is exited.
func (s *BaseJava20ParserListener) ExitInterfaceMethodModifier(ctx *InterfaceMethodModifierContext) {}

// EnterAnnotationInterfaceDeclaration is called when production annotationInterfaceDeclaration is entered.
func (s *BaseJava20ParserListener) EnterAnnotationInterfaceDeclaration(ctx *AnnotationInterfaceDeclarationContext) {
}

// ExitAnnotationInterfaceDeclaration is called when production annotationInterfaceDeclaration is exited.
func (s *BaseJava20ParserListener) ExitAnnotationInterfaceDeclaration(ctx *AnnotationInterfaceDeclarationContext) {
}

// EnterAnnotationInterfaceBody is called when production annotationInterfaceBody is entered.
func (s *BaseJava20ParserListener) EnterAnnotationInterfaceBody(ctx *AnnotationInterfaceBodyContext) {
}

// ExitAnnotationInterfaceBody is called when production annotationInterfaceBody is exited.
func (s *BaseJava20ParserListener) ExitAnnotationInterfaceBody(ctx *AnnotationInterfaceBodyContext) {}

// EnterAnnotationInterfaceMemberDeclaration is called when production annotationInterfaceMemberDeclaration is entered.
func (s *BaseJava20ParserListener) EnterAnnotationInterfaceMemberDeclaration(ctx *AnnotationInterfaceMemberDeclarationContext) {
}

// ExitAnnotationInterfaceMemberDeclaration is called when production annotationInterfaceMemberDeclaration is exited.
func (s *BaseJava20ParserListener) ExitAnnotationInterfaceMemberDeclaration(ctx *AnnotationInterfaceMemberDeclarationContext) {
}

// EnterAnnotationInterfaceElementDeclaration is called when production annotationInterfaceElementDeclaration is entered.
func (s *BaseJava20ParserListener) EnterAnnotationInterfaceElementDeclaration(ctx *AnnotationInterfaceElementDeclarationContext) {
}

// ExitAnnotationInterfaceElementDeclaration is called when production annotationInterfaceElementDeclaration is exited.
func (s *BaseJava20ParserListener) ExitAnnotationInterfaceElementDeclaration(ctx *AnnotationInterfaceElementDeclarationContext) {
}

// EnterAnnotationInterfaceElementModifier is called when production annotationInterfaceElementModifier is entered.
func (s *BaseJava20ParserListener) EnterAnnotationInterfaceElementModifier(ctx *AnnotationInterfaceElementModifierContext) {
}

// ExitAnnotationInterfaceElementModifier is called when production annotationInterfaceElementModifier is exited.
func (s *BaseJava20ParserListener) ExitAnnotationInterfaceElementModifier(ctx *AnnotationInterfaceElementModifierContext) {
}

// EnterDefaultValue is called when production defaultValue is entered.
func (s *BaseJava20ParserListener) EnterDefaultValue(ctx *DefaultValueContext) {}

// ExitDefaultValue is called when production defaultValue is exited.
func (s *BaseJava20ParserListener) ExitDefaultValue(ctx *DefaultValueContext) {}

// EnterAnnotation is called when production annotation is entered.
func (s *BaseJava20ParserListener) EnterAnnotation(ctx *AnnotationContext) {}

// ExitAnnotation is called when production annotation is exited.
func (s *BaseJava20ParserListener) ExitAnnotation(ctx *AnnotationContext) {}

// EnterNormalAnnotation is called when production normalAnnotation is entered.
func (s *BaseJava20ParserListener) EnterNormalAnnotation(ctx *NormalAnnotationContext) {}

// ExitNormalAnnotation is called when production normalAnnotation is exited.
func (s *BaseJava20ParserListener) ExitNormalAnnotation(ctx *NormalAnnotationContext) {}

// EnterElementValuePairList is called when production elementValuePairList is entered.
func (s *BaseJava20ParserListener) EnterElementValuePairList(ctx *ElementValuePairListContext) {}

// ExitElementValuePairList is called when production elementValuePairList is exited.
func (s *BaseJava20ParserListener) ExitElementValuePairList(ctx *ElementValuePairListContext) {}

// EnterElementValuePair is called when production elementValuePair is entered.
func (s *BaseJava20ParserListener) EnterElementValuePair(ctx *ElementValuePairContext) {}

// ExitElementValuePair is called when production elementValuePair is exited.
func (s *BaseJava20ParserListener) ExitElementValuePair(ctx *ElementValuePairContext) {}

// EnterElementValue is called when production elementValue is entered.
func (s *BaseJava20ParserListener) EnterElementValue(ctx *ElementValueContext) {}

// ExitElementValue is called when production elementValue is exited.
func (s *BaseJava20ParserListener) ExitElementValue(ctx *ElementValueContext) {}

// EnterElementValueArrayInitializer is called when production elementValueArrayInitializer is entered.
func (s *BaseJava20ParserListener) EnterElementValueArrayInitializer(ctx *ElementValueArrayInitializerContext) {
}

// ExitElementValueArrayInitializer is called when production elementValueArrayInitializer is exited.
func (s *BaseJava20ParserListener) ExitElementValueArrayInitializer(ctx *ElementValueArrayInitializerContext) {
}

// EnterElementValueList is called when production elementValueList is entered.
func (s *BaseJava20ParserListener) EnterElementValueList(ctx *ElementValueListContext) {}

// ExitElementValueList is called when production elementValueList is exited.
func (s *BaseJava20ParserListener) ExitElementValueList(ctx *ElementValueListContext) {}

// EnterMarkerAnnotation is called when production markerAnnotation is entered.
func (s *BaseJava20ParserListener) EnterMarkerAnnotation(ctx *MarkerAnnotationContext) {}

// ExitMarkerAnnotation is called when production markerAnnotation is exited.
func (s *BaseJava20ParserListener) ExitMarkerAnnotation(ctx *MarkerAnnotationContext) {}

// EnterSingleElementAnnotation is called when production singleElementAnnotation is entered.
func (s *BaseJava20ParserListener) EnterSingleElementAnnotation(ctx *SingleElementAnnotationContext) {
}

// ExitSingleElementAnnotation is called when production singleElementAnnotation is exited.
func (s *BaseJava20ParserListener) ExitSingleElementAnnotation(ctx *SingleElementAnnotationContext) {}

// EnterArrayInitializer is called when production arrayInitializer is entered.
func (s *BaseJava20ParserListener) EnterArrayInitializer(ctx *ArrayInitializerContext) {}

// ExitArrayInitializer is called when production arrayInitializer is exited.
func (s *BaseJava20ParserListener) ExitArrayInitializer(ctx *ArrayInitializerContext) {}

// EnterVariableInitializerList is called when production variableInitializerList is entered.
func (s *BaseJava20ParserListener) EnterVariableInitializerList(ctx *VariableInitializerListContext) {
}

// ExitVariableInitializerList is called when production variableInitializerList is exited.
func (s *BaseJava20ParserListener) ExitVariableInitializerList(ctx *VariableInitializerListContext) {}

// EnterBlock is called when production block is entered.
func (s *BaseJava20ParserListener) EnterBlock(ctx *BlockContext) {}

// ExitBlock is called when production block is exited.
func (s *BaseJava20ParserListener) ExitBlock(ctx *BlockContext) {}

// EnterBlockStatements is called when production blockStatements is entered.
func (s *BaseJava20ParserListener) EnterBlockStatements(ctx *BlockStatementsContext) {}

// ExitBlockStatements is called when production blockStatements is exited.
func (s *BaseJava20ParserListener) ExitBlockStatements(ctx *BlockStatementsContext) {}

// EnterBlockStatement is called when production blockStatement is entered.
func (s *BaseJava20ParserListener) EnterBlockStatement(ctx *BlockStatementContext) {}

// ExitBlockStatement is called when production blockStatement is exited.
func (s *BaseJava20ParserListener) ExitBlockStatement(ctx *BlockStatementContext) {}

// EnterLocalClassOrInterfaceDeclaration is called when production localClassOrInterfaceDeclaration is entered.
func (s *BaseJava20ParserListener) EnterLocalClassOrInterfaceDeclaration(ctx *LocalClassOrInterfaceDeclarationContext) {
}

// ExitLocalClassOrInterfaceDeclaration is called when production localClassOrInterfaceDeclaration is exited.
func (s *BaseJava20ParserListener) ExitLocalClassOrInterfaceDeclaration(ctx *LocalClassOrInterfaceDeclarationContext) {
}

// EnterLocalVariableDeclaration is called when production localVariableDeclaration is entered.
func (s *BaseJava20ParserListener) EnterLocalVariableDeclaration(ctx *LocalVariableDeclarationContext) {
}

// ExitLocalVariableDeclaration is called when production localVariableDeclaration is exited.
func (s *BaseJava20ParserListener) ExitLocalVariableDeclaration(ctx *LocalVariableDeclarationContext) {
}

// EnterLocalVariableType is called when production localVariableType is entered.
func (s *BaseJava20ParserListener) EnterLocalVariableType(ctx *LocalVariableTypeContext) {}

// ExitLocalVariableType is called when production localVariableType is exited.
func (s *BaseJava20ParserListener) ExitLocalVariableType(ctx *LocalVariableTypeContext) {}

// EnterLocalVariableDeclarationStatement is called when production localVariableDeclarationStatement is entered.
func (s *BaseJava20ParserListener) EnterLocalVariableDeclarationStatement(ctx *LocalVariableDeclarationStatementContext) {
}

// ExitLocalVariableDeclarationStatement is called when production localVariableDeclarationStatement is exited.
func (s *BaseJava20ParserListener) ExitLocalVariableDeclarationStatement(ctx *LocalVariableDeclarationStatementContext) {
}

// EnterStatement is called when production statement is entered.
func (s *BaseJava20ParserListener) EnterStatement(ctx *StatementContext) {}

// ExitStatement is called when production statement is exited.
func (s *BaseJava20ParserListener) ExitStatement(ctx *StatementContext) {}

// EnterStatementNoShortIf is called when production statementNoShortIf is entered.
func (s *BaseJava20ParserListener) EnterStatementNoShortIf(ctx *StatementNoShortIfContext) {}

// ExitStatementNoShortIf is called when production statementNoShortIf is exited.
func (s *BaseJava20ParserListener) ExitStatementNoShortIf(ctx *StatementNoShortIfContext) {}

// EnterStatementWithoutTrailingSubstatement is called when production statementWithoutTrailingSubstatement is entered.
func (s *BaseJava20ParserListener) EnterStatementWithoutTrailingSubstatement(ctx *StatementWithoutTrailingSubstatementContext) {
}

// ExitStatementWithoutTrailingSubstatement is called when production statementWithoutTrailingSubstatement is exited.
func (s *BaseJava20ParserListener) ExitStatementWithoutTrailingSubstatement(ctx *StatementWithoutTrailingSubstatementContext) {
}

// EnterEmptyStatement_ is called when production emptyStatement_ is entered.
func (s *BaseJava20ParserListener) EnterEmptyStatement_(ctx *EmptyStatement_Context) {}

// ExitEmptyStatement_ is called when production emptyStatement_ is exited.
func (s *BaseJava20ParserListener) ExitEmptyStatement_(ctx *EmptyStatement_Context) {}

// EnterLabeledStatement is called when production labeledStatement is entered.
func (s *BaseJava20ParserListener) EnterLabeledStatement(ctx *LabeledStatementContext) {}

// ExitLabeledStatement is called when production labeledStatement is exited.
func (s *BaseJava20ParserListener) ExitLabeledStatement(ctx *LabeledStatementContext) {}

// EnterLabeledStatementNoShortIf is called when production labeledStatementNoShortIf is entered.
func (s *BaseJava20ParserListener) EnterLabeledStatementNoShortIf(ctx *LabeledStatementNoShortIfContext) {
}

// ExitLabeledStatementNoShortIf is called when production labeledStatementNoShortIf is exited.
func (s *BaseJava20ParserListener) ExitLabeledStatementNoShortIf(ctx *LabeledStatementNoShortIfContext) {
}

// EnterExpressionStatement is called when production expressionStatement is entered.
func (s *BaseJava20ParserListener) EnterExpressionStatement(ctx *ExpressionStatementContext) {}

// ExitExpressionStatement is called when production expressionStatement is exited.
func (s *BaseJava20ParserListener) ExitExpressionStatement(ctx *ExpressionStatementContext) {}

// EnterStatementExpression is called when production statementExpression is entered.
func (s *BaseJava20ParserListener) EnterStatementExpression(ctx *StatementExpressionContext) {}

// ExitStatementExpression is called when production statementExpression is exited.
func (s *BaseJava20ParserListener) ExitStatementExpression(ctx *StatementExpressionContext) {}

// EnterIfThenStatement is called when production ifThenStatement is entered.
func (s *BaseJava20ParserListener) EnterIfThenStatement(ctx *IfThenStatementContext) {}

// ExitIfThenStatement is called when production ifThenStatement is exited.
func (s *BaseJava20ParserListener) ExitIfThenStatement(ctx *IfThenStatementContext) {}

// EnterIfThenElseStatement is called when production ifThenElseStatement is entered.
func (s *BaseJava20ParserListener) EnterIfThenElseStatement(ctx *IfThenElseStatementContext) {}

// ExitIfThenElseStatement is called when production ifThenElseStatement is exited.
func (s *BaseJava20ParserListener) ExitIfThenElseStatement(ctx *IfThenElseStatementContext) {}

// EnterIfThenElseStatementNoShortIf is called when production ifThenElseStatementNoShortIf is entered.
func (s *BaseJava20ParserListener) EnterIfThenElseStatementNoShortIf(ctx *IfThenElseStatementNoShortIfContext) {
}

// ExitIfThenElseStatementNoShortIf is called when production ifThenElseStatementNoShortIf is exited.
func (s *BaseJava20ParserListener) ExitIfThenElseStatementNoShortIf(ctx *IfThenElseStatementNoShortIfContext) {
}

// EnterAssertStatement is called when production assertStatement is entered.
func (s *BaseJava20ParserListener) EnterAssertStatement(ctx *AssertStatementContext) {}

// ExitAssertStatement is called when production assertStatement is exited.
func (s *BaseJava20ParserListener) ExitAssertStatement(ctx *AssertStatementContext) {}

// EnterSwitchStatement is called when production switchStatement is entered.
func (s *BaseJava20ParserListener) EnterSwitchStatement(ctx *SwitchStatementContext) {}

// ExitSwitchStatement is called when production switchStatement is exited.
func (s *BaseJava20ParserListener) ExitSwitchStatement(ctx *SwitchStatementContext) {}

// EnterSwitchBlock is called when production switchBlock is entered.
func (s *BaseJava20ParserListener) EnterSwitchBlock(ctx *SwitchBlockContext) {}

// ExitSwitchBlock is called when production switchBlock is exited.
func (s *BaseJava20ParserListener) ExitSwitchBlock(ctx *SwitchBlockContext) {}

// EnterSwitchRule is called when production switchRule is entered.
func (s *BaseJava20ParserListener) EnterSwitchRule(ctx *SwitchRuleContext) {}

// ExitSwitchRule is called when production switchRule is exited.
func (s *BaseJava20ParserListener) ExitSwitchRule(ctx *SwitchRuleContext) {}

// EnterSwitchBlockStatementGroup is called when production switchBlockStatementGroup is entered.
func (s *BaseJava20ParserListener) EnterSwitchBlockStatementGroup(ctx *SwitchBlockStatementGroupContext) {
}

// ExitSwitchBlockStatementGroup is called when production switchBlockStatementGroup is exited.
func (s *BaseJava20ParserListener) ExitSwitchBlockStatementGroup(ctx *SwitchBlockStatementGroupContext) {
}

// EnterSwitchLabel is called when production switchLabel is entered.
func (s *BaseJava20ParserListener) EnterSwitchLabel(ctx *SwitchLabelContext) {}

// ExitSwitchLabel is called when production switchLabel is exited.
func (s *BaseJava20ParserListener) ExitSwitchLabel(ctx *SwitchLabelContext) {}

// EnterCaseConstant is called when production caseConstant is entered.
func (s *BaseJava20ParserListener) EnterCaseConstant(ctx *CaseConstantContext) {}

// ExitCaseConstant is called when production caseConstant is exited.
func (s *BaseJava20ParserListener) ExitCaseConstant(ctx *CaseConstantContext) {}

// EnterWhileStatement is called when production whileStatement is entered.
func (s *BaseJava20ParserListener) EnterWhileStatement(ctx *WhileStatementContext) {}

// ExitWhileStatement is called when production whileStatement is exited.
func (s *BaseJava20ParserListener) ExitWhileStatement(ctx *WhileStatementContext) {}

// EnterWhileStatementNoShortIf is called when production whileStatementNoShortIf is entered.
func (s *BaseJava20ParserListener) EnterWhileStatementNoShortIf(ctx *WhileStatementNoShortIfContext) {
}

// ExitWhileStatementNoShortIf is called when production whileStatementNoShortIf is exited.
func (s *BaseJava20ParserListener) ExitWhileStatementNoShortIf(ctx *WhileStatementNoShortIfContext) {}

// EnterDoStatement is called when production doStatement is entered.
func (s *BaseJava20ParserListener) EnterDoStatement(ctx *DoStatementContext) {}

// ExitDoStatement is called when production doStatement is exited.
func (s *BaseJava20ParserListener) ExitDoStatement(ctx *DoStatementContext) {}

// EnterForStatement is called when production forStatement is entered.
func (s *BaseJava20ParserListener) EnterForStatement(ctx *ForStatementContext) {}

// ExitForStatement is called when production forStatement is exited.
func (s *BaseJava20ParserListener) ExitForStatement(ctx *ForStatementContext) {}

// EnterForStatementNoShortIf is called when production forStatementNoShortIf is entered.
func (s *BaseJava20ParserListener) EnterForStatementNoShortIf(ctx *ForStatementNoShortIfContext) {}

// ExitForStatementNoShortIf is called when production forStatementNoShortIf is exited.
func (s *BaseJava20ParserListener) ExitForStatementNoShortIf(ctx *ForStatementNoShortIfContext) {}

// EnterBasicForStatement is called when production basicForStatement is entered.
func (s *BaseJava20ParserListener) EnterBasicForStatement(ctx *BasicForStatementContext) {}

// ExitBasicForStatement is called when production basicForStatement is exited.
func (s *BaseJava20ParserListener) ExitBasicForStatement(ctx *BasicForStatementContext) {}

// EnterBasicForStatementNoShortIf is called when production basicForStatementNoShortIf is entered.
func (s *BaseJava20ParserListener) EnterBasicForStatementNoShortIf(ctx *BasicForStatementNoShortIfContext) {
}

// ExitBasicForStatementNoShortIf is called when production basicForStatementNoShortIf is exited.
func (s *BaseJava20ParserListener) ExitBasicForStatementNoShortIf(ctx *BasicForStatementNoShortIfContext) {
}

// EnterForInit is called when production forInit is entered.
func (s *BaseJava20ParserListener) EnterForInit(ctx *ForInitContext) {}

// ExitForInit is called when production forInit is exited.
func (s *BaseJava20ParserListener) ExitForInit(ctx *ForInitContext) {}

// EnterForUpdate is called when production forUpdate is entered.
func (s *BaseJava20ParserListener) EnterForUpdate(ctx *ForUpdateContext) {}

// ExitForUpdate is called when production forUpdate is exited.
func (s *BaseJava20ParserListener) ExitForUpdate(ctx *ForUpdateContext) {}

// EnterStatementExpressionList is called when production statementExpressionList is entered.
func (s *BaseJava20ParserListener) EnterStatementExpressionList(ctx *StatementExpressionListContext) {
}

// ExitStatementExpressionList is called when production statementExpressionList is exited.
func (s *BaseJava20ParserListener) ExitStatementExpressionList(ctx *StatementExpressionListContext) {}

// EnterEnhancedForStatement is called when production enhancedForStatement is entered.
func (s *BaseJava20ParserListener) EnterEnhancedForStatement(ctx *EnhancedForStatementContext) {}

// ExitEnhancedForStatement is called when production enhancedForStatement is exited.
func (s *BaseJava20ParserListener) ExitEnhancedForStatement(ctx *EnhancedForStatementContext) {}

// EnterEnhancedForStatementNoShortIf is called when production enhancedForStatementNoShortIf is entered.
func (s *BaseJava20ParserListener) EnterEnhancedForStatementNoShortIf(ctx *EnhancedForStatementNoShortIfContext) {
}

// ExitEnhancedForStatementNoShortIf is called when production enhancedForStatementNoShortIf is exited.
func (s *BaseJava20ParserListener) ExitEnhancedForStatementNoShortIf(ctx *EnhancedForStatementNoShortIfContext) {
}

// EnterBreakStatement is called when production breakStatement is entered.
func (s *BaseJava20ParserListener) EnterBreakStatement(ctx *BreakStatementContext) {}

// ExitBreakStatement is called when production breakStatement is exited.
func (s *BaseJava20ParserListener) ExitBreakStatement(ctx *BreakStatementContext) {}

// EnterContinueStatement is called when production continueStatement is entered.
func (s *BaseJava20ParserListener) EnterContinueStatement(ctx *ContinueStatementContext) {}

// ExitContinueStatement is called when production continueStatement is exited.
func (s *BaseJava20ParserListener) ExitContinueStatement(ctx *ContinueStatementContext) {}

// EnterReturnStatement is called when production returnStatement is entered.
func (s *BaseJava20ParserListener) EnterReturnStatement(ctx *ReturnStatementContext) {}

// ExitReturnStatement is called when production returnStatement is exited.
func (s *BaseJava20ParserListener) ExitReturnStatement(ctx *ReturnStatementContext) {}

// EnterThrowStatement is called when production throwStatement is entered.
func (s *BaseJava20ParserListener) EnterThrowStatement(ctx *ThrowStatementContext) {}

// ExitThrowStatement is called when production throwStatement is exited.
func (s *BaseJava20ParserListener) ExitThrowStatement(ctx *ThrowStatementContext) {}

// EnterSynchronizedStatement is called when production synchronizedStatement is entered.
func (s *BaseJava20ParserListener) EnterSynchronizedStatement(ctx *SynchronizedStatementContext) {}

// ExitSynchronizedStatement is called when production synchronizedStatement is exited.
func (s *BaseJava20ParserListener) ExitSynchronizedStatement(ctx *SynchronizedStatementContext) {}

// EnterTryStatement is called when production tryStatement is entered.
func (s *BaseJava20ParserListener) EnterTryStatement(ctx *TryStatementContext) {}

// ExitTryStatement is called when production tryStatement is exited.
func (s *BaseJava20ParserListener) ExitTryStatement(ctx *TryStatementContext) {}

// EnterCatches is called when production catches is entered.
func (s *BaseJava20ParserListener) EnterCatches(ctx *CatchesContext) {}

// ExitCatches is called when production catches is exited.
func (s *BaseJava20ParserListener) ExitCatches(ctx *CatchesContext) {}

// EnterCatchClause is called when production catchClause is entered.
func (s *BaseJava20ParserListener) EnterCatchClause(ctx *CatchClauseContext) {}

// ExitCatchClause is called when production catchClause is exited.
func (s *BaseJava20ParserListener) ExitCatchClause(ctx *CatchClauseContext) {}

// EnterCatchFormalParameter is called when production catchFormalParameter is entered.
func (s *BaseJava20ParserListener) EnterCatchFormalParameter(ctx *CatchFormalParameterContext) {}

// ExitCatchFormalParameter is called when production catchFormalParameter is exited.
func (s *BaseJava20ParserListener) ExitCatchFormalParameter(ctx *CatchFormalParameterContext) {}

// EnterCatchType is called when production catchType is entered.
func (s *BaseJava20ParserListener) EnterCatchType(ctx *CatchTypeContext) {}

// ExitCatchType is called when production catchType is exited.
func (s *BaseJava20ParserListener) ExitCatchType(ctx *CatchTypeContext) {}

// EnterFinallyBlock is called when production finallyBlock is entered.
func (s *BaseJava20ParserListener) EnterFinallyBlock(ctx *FinallyBlockContext) {}

// ExitFinallyBlock is called when production finallyBlock is exited.
func (s *BaseJava20ParserListener) ExitFinallyBlock(ctx *FinallyBlockContext) {}

// EnterTryWithResourcesStatement is called when production tryWithResourcesStatement is entered.
func (s *BaseJava20ParserListener) EnterTryWithResourcesStatement(ctx *TryWithResourcesStatementContext) {
}

// ExitTryWithResourcesStatement is called when production tryWithResourcesStatement is exited.
func (s *BaseJava20ParserListener) ExitTryWithResourcesStatement(ctx *TryWithResourcesStatementContext) {
}

// EnterResourceSpecification is called when production resourceSpecification is entered.
func (s *BaseJava20ParserListener) EnterResourceSpecification(ctx *ResourceSpecificationContext) {}

// ExitResourceSpecification is called when production resourceSpecification is exited.
func (s *BaseJava20ParserListener) ExitResourceSpecification(ctx *ResourceSpecificationContext) {}

// EnterResourceList is called when production resourceList is entered.
func (s *BaseJava20ParserListener) EnterResourceList(ctx *ResourceListContext) {}

// ExitResourceList is called when production resourceList is exited.
func (s *BaseJava20ParserListener) ExitResourceList(ctx *ResourceListContext) {}

// EnterResource is called when production resource is entered.
func (s *BaseJava20ParserListener) EnterResource(ctx *ResourceContext) {}

// ExitResource is called when production resource is exited.
func (s *BaseJava20ParserListener) ExitResource(ctx *ResourceContext) {}

// EnterVariableAccess is called when production variableAccess is entered.
func (s *BaseJava20ParserListener) EnterVariableAccess(ctx *VariableAccessContext) {}

// ExitVariableAccess is called when production variableAccess is exited.
func (s *BaseJava20ParserListener) ExitVariableAccess(ctx *VariableAccessContext) {}

// EnterYieldStatement is called when production yieldStatement is entered.
func (s *BaseJava20ParserListener) EnterYieldStatement(ctx *YieldStatementContext) {}

// ExitYieldStatement is called when production yieldStatement is exited.
func (s *BaseJava20ParserListener) ExitYieldStatement(ctx *YieldStatementContext) {}

// EnterPattern is called when production pattern is entered.
func (s *BaseJava20ParserListener) EnterPattern(ctx *PatternContext) {}

// ExitPattern is called when production pattern is exited.
func (s *BaseJava20ParserListener) ExitPattern(ctx *PatternContext) {}

// EnterTypePattern is called when production typePattern is entered.
func (s *BaseJava20ParserListener) EnterTypePattern(ctx *TypePatternContext) {}

// ExitTypePattern is called when production typePattern is exited.
func (s *BaseJava20ParserListener) ExitTypePattern(ctx *TypePatternContext) {}

// EnterExpression is called when production expression is entered.
func (s *BaseJava20ParserListener) EnterExpression(ctx *ExpressionContext) {}

// ExitExpression is called when production expression is exited.
func (s *BaseJava20ParserListener) ExitExpression(ctx *ExpressionContext) {}

// EnterPrimary is called when production primary is entered.
func (s *BaseJava20ParserListener) EnterPrimary(ctx *PrimaryContext) {}

// ExitPrimary is called when production primary is exited.
func (s *BaseJava20ParserListener) ExitPrimary(ctx *PrimaryContext) {}

// EnterPrimaryNoNewArray is called when production primaryNoNewArray is entered.
func (s *BaseJava20ParserListener) EnterPrimaryNoNewArray(ctx *PrimaryNoNewArrayContext) {}

// ExitPrimaryNoNewArray is called when production primaryNoNewArray is exited.
func (s *BaseJava20ParserListener) ExitPrimaryNoNewArray(ctx *PrimaryNoNewArrayContext) {}

// EnterPNNA is called when production pNNA is entered.
func (s *BaseJava20ParserListener) EnterPNNA(ctx *PNNAContext) {}

// ExitPNNA is called when production pNNA is exited.
func (s *BaseJava20ParserListener) ExitPNNA(ctx *PNNAContext) {}

// EnterClassLiteral is called when production classLiteral is entered.
func (s *BaseJava20ParserListener) EnterClassLiteral(ctx *ClassLiteralContext) {}

// ExitClassLiteral is called when production classLiteral is exited.
func (s *BaseJava20ParserListener) ExitClassLiteral(ctx *ClassLiteralContext) {}

// EnterClassInstanceCreationExpression is called when production classInstanceCreationExpression is entered.
func (s *BaseJava20ParserListener) EnterClassInstanceCreationExpression(ctx *ClassInstanceCreationExpressionContext) {
}

// ExitClassInstanceCreationExpression is called when production classInstanceCreationExpression is exited.
func (s *BaseJava20ParserListener) ExitClassInstanceCreationExpression(ctx *ClassInstanceCreationExpressionContext) {
}

// EnterUnqualifiedClassInstanceCreationExpression is called when production unqualifiedClassInstanceCreationExpression is entered.
func (s *BaseJava20ParserListener) EnterUnqualifiedClassInstanceCreationExpression(ctx *UnqualifiedClassInstanceCreationExpressionContext) {
}

// ExitUnqualifiedClassInstanceCreationExpression is called when production unqualifiedClassInstanceCreationExpression is exited.
func (s *BaseJava20ParserListener) ExitUnqualifiedClassInstanceCreationExpression(ctx *UnqualifiedClassInstanceCreationExpressionContext) {
}

// EnterClassOrInterfaceTypeToInstantiate is called when production classOrInterfaceTypeToInstantiate is entered.
func (s *BaseJava20ParserListener) EnterClassOrInterfaceTypeToInstantiate(ctx *ClassOrInterfaceTypeToInstantiateContext) {
}

// ExitClassOrInterfaceTypeToInstantiate is called when production classOrInterfaceTypeToInstantiate is exited.
func (s *BaseJava20ParserListener) ExitClassOrInterfaceTypeToInstantiate(ctx *ClassOrInterfaceTypeToInstantiateContext) {
}

// EnterTypeArgumentsOrDiamond is called when production typeArgumentsOrDiamond is entered.
func (s *BaseJava20ParserListener) EnterTypeArgumentsOrDiamond(ctx *TypeArgumentsOrDiamondContext) {}

// ExitTypeArgumentsOrDiamond is called when production typeArgumentsOrDiamond is exited.
func (s *BaseJava20ParserListener) ExitTypeArgumentsOrDiamond(ctx *TypeArgumentsOrDiamondContext) {}

// EnterArrayCreationExpression is called when production arrayCreationExpression is entered.
func (s *BaseJava20ParserListener) EnterArrayCreationExpression(ctx *ArrayCreationExpressionContext) {
}

// ExitArrayCreationExpression is called when production arrayCreationExpression is exited.
func (s *BaseJava20ParserListener) ExitArrayCreationExpression(ctx *ArrayCreationExpressionContext) {}

// EnterArrayCreationExpressionWithoutInitializer is called when production arrayCreationExpressionWithoutInitializer is entered.
func (s *BaseJava20ParserListener) EnterArrayCreationExpressionWithoutInitializer(ctx *ArrayCreationExpressionWithoutInitializerContext) {
}

// ExitArrayCreationExpressionWithoutInitializer is called when production arrayCreationExpressionWithoutInitializer is exited.
func (s *BaseJava20ParserListener) ExitArrayCreationExpressionWithoutInitializer(ctx *ArrayCreationExpressionWithoutInitializerContext) {
}

// EnterArrayCreationExpressionWithInitializer is called when production arrayCreationExpressionWithInitializer is entered.
func (s *BaseJava20ParserListener) EnterArrayCreationExpressionWithInitializer(ctx *ArrayCreationExpressionWithInitializerContext) {
}

// ExitArrayCreationExpressionWithInitializer is called when production arrayCreationExpressionWithInitializer is exited.
func (s *BaseJava20ParserListener) ExitArrayCreationExpressionWithInitializer(ctx *ArrayCreationExpressionWithInitializerContext) {
}

// EnterDimExprs is called when production dimExprs is entered.
func (s *BaseJava20ParserListener) EnterDimExprs(ctx *DimExprsContext) {}

// ExitDimExprs is called when production dimExprs is exited.
func (s *BaseJava20ParserListener) ExitDimExprs(ctx *DimExprsContext) {}

// EnterDimExpr is called when production dimExpr is entered.
func (s *BaseJava20ParserListener) EnterDimExpr(ctx *DimExprContext) {}

// ExitDimExpr is called when production dimExpr is exited.
func (s *BaseJava20ParserListener) ExitDimExpr(ctx *DimExprContext) {}

// EnterArrayAccess is called when production arrayAccess is entered.
func (s *BaseJava20ParserListener) EnterArrayAccess(ctx *ArrayAccessContext) {}

// ExitArrayAccess is called when production arrayAccess is exited.
func (s *BaseJava20ParserListener) ExitArrayAccess(ctx *ArrayAccessContext) {}

// EnterFieldAccess is called when production fieldAccess is entered.
func (s *BaseJava20ParserListener) EnterFieldAccess(ctx *FieldAccessContext) {}

// ExitFieldAccess is called when production fieldAccess is exited.
func (s *BaseJava20ParserListener) ExitFieldAccess(ctx *FieldAccessContext) {}

// EnterMethodInvocation is called when production methodInvocation is entered.
func (s *BaseJava20ParserListener) EnterMethodInvocation(ctx *MethodInvocationContext) {}

// ExitMethodInvocation is called when production methodInvocation is exited.
func (s *BaseJava20ParserListener) ExitMethodInvocation(ctx *MethodInvocationContext) {}

// EnterArgumentList is called when production argumentList is entered.
func (s *BaseJava20ParserListener) EnterArgumentList(ctx *ArgumentListContext) {}

// ExitArgumentList is called when production argumentList is exited.
func (s *BaseJava20ParserListener) ExitArgumentList(ctx *ArgumentListContext) {}

// EnterMethodReference is called when production methodReference is entered.
func (s *BaseJava20ParserListener) EnterMethodReference(ctx *MethodReferenceContext) {}

// ExitMethodReference is called when production methodReference is exited.
func (s *BaseJava20ParserListener) ExitMethodReference(ctx *MethodReferenceContext) {}

// EnterPostfixExpression is called when production postfixExpression is entered.
func (s *BaseJava20ParserListener) EnterPostfixExpression(ctx *PostfixExpressionContext) {}

// ExitPostfixExpression is called when production postfixExpression is exited.
func (s *BaseJava20ParserListener) ExitPostfixExpression(ctx *PostfixExpressionContext) {}

// EnterPfE is called when production pfE is entered.
func (s *BaseJava20ParserListener) EnterPfE(ctx *PfEContext) {}

// ExitPfE is called when production pfE is exited.
func (s *BaseJava20ParserListener) ExitPfE(ctx *PfEContext) {}

// EnterPostIncrementExpression is called when production postIncrementExpression is entered.
func (s *BaseJava20ParserListener) EnterPostIncrementExpression(ctx *PostIncrementExpressionContext) {
}

// ExitPostIncrementExpression is called when production postIncrementExpression is exited.
func (s *BaseJava20ParserListener) ExitPostIncrementExpression(ctx *PostIncrementExpressionContext) {}

// EnterPostDecrementExpression is called when production postDecrementExpression is entered.
func (s *BaseJava20ParserListener) EnterPostDecrementExpression(ctx *PostDecrementExpressionContext) {
}

// ExitPostDecrementExpression is called when production postDecrementExpression is exited.
func (s *BaseJava20ParserListener) ExitPostDecrementExpression(ctx *PostDecrementExpressionContext) {}

// EnterUnaryExpression is called when production unaryExpression is entered.
func (s *BaseJava20ParserListener) EnterUnaryExpression(ctx *UnaryExpressionContext) {}

// ExitUnaryExpression is called when production unaryExpression is exited.
func (s *BaseJava20ParserListener) ExitUnaryExpression(ctx *UnaryExpressionContext) {}

// EnterPreIncrementExpression is called when production preIncrementExpression is entered.
func (s *BaseJava20ParserListener) EnterPreIncrementExpression(ctx *PreIncrementExpressionContext) {}

// ExitPreIncrementExpression is called when production preIncrementExpression is exited.
func (s *BaseJava20ParserListener) ExitPreIncrementExpression(ctx *PreIncrementExpressionContext) {}

// EnterPreDecrementExpression is called when production preDecrementExpression is entered.
func (s *BaseJava20ParserListener) EnterPreDecrementExpression(ctx *PreDecrementExpressionContext) {}

// ExitPreDecrementExpression is called when production preDecrementExpression is exited.
func (s *BaseJava20ParserListener) ExitPreDecrementExpression(ctx *PreDecrementExpressionContext) {}

// EnterUnaryExpressionNotPlusMinus is called when production unaryExpressionNotPlusMinus is entered.
func (s *BaseJava20ParserListener) EnterUnaryExpressionNotPlusMinus(ctx *UnaryExpressionNotPlusMinusContext) {
}

// ExitUnaryExpressionNotPlusMinus is called when production unaryExpressionNotPlusMinus is exited.
func (s *BaseJava20ParserListener) ExitUnaryExpressionNotPlusMinus(ctx *UnaryExpressionNotPlusMinusContext) {
}

// EnterCastExpression is called when production castExpression is entered.
func (s *BaseJava20ParserListener) EnterCastExpression(ctx *CastExpressionContext) {}

// ExitCastExpression is called when production castExpression is exited.
func (s *BaseJava20ParserListener) ExitCastExpression(ctx *CastExpressionContext) {}

// EnterMultiplicativeExpression is called when production multiplicativeExpression is entered.
func (s *BaseJava20ParserListener) EnterMultiplicativeExpression(ctx *MultiplicativeExpressionContext) {
}

// ExitMultiplicativeExpression is called when production multiplicativeExpression is exited.
func (s *BaseJava20ParserListener) ExitMultiplicativeExpression(ctx *MultiplicativeExpressionContext) {
}

// EnterAdditiveExpression is called when production additiveExpression is entered.
func (s *BaseJava20ParserListener) EnterAdditiveExpression(ctx *AdditiveExpressionContext) {}

// ExitAdditiveExpression is called when production additiveExpression is exited.
func (s *BaseJava20ParserListener) ExitAdditiveExpression(ctx *AdditiveExpressionContext) {}

// EnterShiftExpression is called when production shiftExpression is entered.
func (s *BaseJava20ParserListener) EnterShiftExpression(ctx *ShiftExpressionContext) {}

// ExitShiftExpression is called when production shiftExpression is exited.
func (s *BaseJava20ParserListener) ExitShiftExpression(ctx *ShiftExpressionContext) {}

// EnterRelationalExpression is called when production relationalExpression is entered.
func (s *BaseJava20ParserListener) EnterRelationalExpression(ctx *RelationalExpressionContext) {}

// ExitRelationalExpression is called when production relationalExpression is exited.
func (s *BaseJava20ParserListener) ExitRelationalExpression(ctx *RelationalExpressionContext) {}

// EnterEqualityExpression is called when production equalityExpression is entered.
func (s *BaseJava20ParserListener) EnterEqualityExpression(ctx *EqualityExpressionContext) {}

// ExitEqualityExpression is called when production equalityExpression is exited.
func (s *BaseJava20ParserListener) ExitEqualityExpression(ctx *EqualityExpressionContext) {}

// EnterAndExpression is called when production andExpression is entered.
func (s *BaseJava20ParserListener) EnterAndExpression(ctx *AndExpressionContext) {}

// ExitAndExpression is called when production andExpression is exited.
func (s *BaseJava20ParserListener) ExitAndExpression(ctx *AndExpressionContext) {}

// EnterExclusiveOrExpression is called when production exclusiveOrExpression is entered.
func (s *BaseJava20ParserListener) EnterExclusiveOrExpression(ctx *ExclusiveOrExpressionContext) {}

// ExitExclusiveOrExpression is called when production exclusiveOrExpression is exited.
func (s *BaseJava20ParserListener) ExitExclusiveOrExpression(ctx *ExclusiveOrExpressionContext) {}

// EnterInclusiveOrExpression is called when production inclusiveOrExpression is entered.
func (s *BaseJava20ParserListener) EnterInclusiveOrExpression(ctx *InclusiveOrExpressionContext) {}

// ExitInclusiveOrExpression is called when production inclusiveOrExpression is exited.
func (s *BaseJava20ParserListener) ExitInclusiveOrExpression(ctx *InclusiveOrExpressionContext) {}

// EnterConditionalAndExpression is called when production conditionalAndExpression is entered.
func (s *BaseJava20ParserListener) EnterConditionalAndExpression(ctx *ConditionalAndExpressionContext) {
}

// ExitConditionalAndExpression is called when production conditionalAndExpression is exited.
func (s *BaseJava20ParserListener) ExitConditionalAndExpression(ctx *ConditionalAndExpressionContext) {
}

// EnterConditionalOrExpression is called when production conditionalOrExpression is entered.
func (s *BaseJava20ParserListener) EnterConditionalOrExpression(ctx *ConditionalOrExpressionContext) {
}

// ExitConditionalOrExpression is called when production conditionalOrExpression is exited.
func (s *BaseJava20ParserListener) ExitConditionalOrExpression(ctx *ConditionalOrExpressionContext) {}

// EnterConditionalExpression is called when production conditionalExpression is entered.
func (s *BaseJava20ParserListener) EnterConditionalExpression(ctx *ConditionalExpressionContext) {}

// ExitConditionalExpression is called when production conditionalExpression is exited.
func (s *BaseJava20ParserListener) ExitConditionalExpression(ctx *ConditionalExpressionContext) {}

// EnterAssignmentExpression is called when production assignmentExpression is entered.
func (s *BaseJava20ParserListener) EnterAssignmentExpression(ctx *AssignmentExpressionContext) {}

// ExitAssignmentExpression is called when production assignmentExpression is exited.
func (s *BaseJava20ParserListener) ExitAssignmentExpression(ctx *AssignmentExpressionContext) {}

// EnterAssignment is called when production assignment is entered.
func (s *BaseJava20ParserListener) EnterAssignment(ctx *AssignmentContext) {}

// ExitAssignment is called when production assignment is exited.
func (s *BaseJava20ParserListener) ExitAssignment(ctx *AssignmentContext) {}

// EnterLeftHandSide is called when production leftHandSide is entered.
func (s *BaseJava20ParserListener) EnterLeftHandSide(ctx *LeftHandSideContext) {}

// ExitLeftHandSide is called when production leftHandSide is exited.
func (s *BaseJava20ParserListener) ExitLeftHandSide(ctx *LeftHandSideContext) {}

// EnterAssignmentOperator is called when production assignmentOperator is entered.
func (s *BaseJava20ParserListener) EnterAssignmentOperator(ctx *AssignmentOperatorContext) {}

// ExitAssignmentOperator is called when production assignmentOperator is exited.
func (s *BaseJava20ParserListener) ExitAssignmentOperator(ctx *AssignmentOperatorContext) {}

// EnterLambdaExpression is called when production lambdaExpression is entered.
func (s *BaseJava20ParserListener) EnterLambdaExpression(ctx *LambdaExpressionContext) {}

// ExitLambdaExpression is called when production lambdaExpression is exited.
func (s *BaseJava20ParserListener) ExitLambdaExpression(ctx *LambdaExpressionContext) {}

// EnterLambdaParameters is called when production lambdaParameters is entered.
func (s *BaseJava20ParserListener) EnterLambdaParameters(ctx *LambdaParametersContext) {}

// ExitLambdaParameters is called when production lambdaParameters is exited.
func (s *BaseJava20ParserListener) ExitLambdaParameters(ctx *LambdaParametersContext) {}

// EnterLambdaParameterList is called when production lambdaParameterList is entered.
func (s *BaseJava20ParserListener) EnterLambdaParameterList(ctx *LambdaParameterListContext) {}

// ExitLambdaParameterList is called when production lambdaParameterList is exited.
func (s *BaseJava20ParserListener) ExitLambdaParameterList(ctx *LambdaParameterListContext) {}

// EnterLambdaParameter is called when production lambdaParameter is entered.
func (s *BaseJava20ParserListener) EnterLambdaParameter(ctx *LambdaParameterContext) {}

// ExitLambdaParameter is called when production lambdaParameter is exited.
func (s *BaseJava20ParserListener) ExitLambdaParameter(ctx *LambdaParameterContext) {}

// EnterLambdaParameterType is called when production lambdaParameterType is entered.
func (s *BaseJava20ParserListener) EnterLambdaParameterType(ctx *LambdaParameterTypeContext) {}

// ExitLambdaParameterType is called when production lambdaParameterType is exited.
func (s *BaseJava20ParserListener) ExitLambdaParameterType(ctx *LambdaParameterTypeContext) {}

// EnterLambdaBody is called when production lambdaBody is entered.
func (s *BaseJava20ParserListener) EnterLambdaBody(ctx *LambdaBodyContext) {}

// ExitLambdaBody is called when production lambdaBody is exited.
func (s *BaseJava20ParserListener) ExitLambdaBody(ctx *LambdaBodyContext) {}

// EnterSwitchExpression is called when production switchExpression is entered.
func (s *BaseJava20ParserListener) EnterSwitchExpression(ctx *SwitchExpressionContext) {}

// ExitSwitchExpression is called when production switchExpression is exited.
func (s *BaseJava20ParserListener) ExitSwitchExpression(ctx *SwitchExpressionContext) {}

// EnterConstantExpression is called when production constantExpression is entered.
func (s *BaseJava20ParserListener) EnterConstantExpression(ctx *ConstantExpressionContext) {}

// ExitConstantExpression is called when production constantExpression is exited.
func (s *BaseJava20ParserListener) ExitConstantExpression(ctx *ConstantExpressionContext) {}
