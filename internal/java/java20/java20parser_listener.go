// Code generated from Java20Parser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package java20 // Java20Parser
import "github.com/antlr4-go/antlr/v4"

// Java20ParserListener is a complete listener for a parse tree produced by Java20Parser.
type Java20ParserListener interface {
	antlr.ParseTreeListener

	// EnterStart_ is called when entering the start_ production.
	EnterStart_(c *Start_Context)

	// EnterIdentifier is called when entering the identifier production.
	EnterIdentifier(c *IdentifierContext)

	// EnterTypeIdentifier is called when entering the typeIdentifier production.
	EnterTypeIdentifier(c *TypeIdentifierContext)

	// EnterUnqualifiedMethodIdentifier is called when entering the unqualifiedMethodIdentifier production.
	EnterUnqualifiedMethodIdentifier(c *UnqualifiedMethodIdentifierContext)

	// EnterContextualKeyword is called when entering the contextualKeyword production.
	EnterContextualKeyword(c *ContextualKeywordContext)

	// EnterContextualKeywordMinusForTypeIdentifier is called when entering the contextualKeywordMinusForTypeIdentifier production.
	EnterContextualKeywordMinusForTypeIdentifier(c *ContextualKeywordMinusForTypeIdentifierContext)

	// EnterContextualKeywordMinusForUnqualifiedMethodIdentifier is called when entering the contextualKeywordMinusForUnqualifiedMethodIdentifier production.
	EnterContextualKeywordMinusForUnqualifiedMethodIdentifier(c *ContextualKeywordMinusForUnqualifiedMethodIdentifierContext)

	// EnterLiteral is called when entering the literal production.
	EnterLiteral(c *LiteralContext)

	// EnterPrimitiveType is called when entering the primitiveType production.
	EnterPrimitiveType(c *PrimitiveTypeContext)

	// EnterNumericType is called when entering the numericType production.
	EnterNumericType(c *NumericTypeContext)

	// EnterIntegralType is called when entering the integralType production.
	EnterIntegralType(c *IntegralTypeContext)

	// EnterFloatingPointType is called when entering the floatingPointType production.
	EnterFloatingPointType(c *FloatingPointTypeContext)

	// EnterReferenceType is called when entering the referenceType production.
	EnterReferenceType(c *ReferenceTypeContext)

	// EnterCoit is called when entering the coit production.
	EnterCoit(c *CoitContext)

	// EnterClassOrInterfaceType is called when entering the classOrInterfaceType production.
	EnterClassOrInterfaceType(c *ClassOrInterfaceTypeContext)

	// EnterClassType is called when entering the classType production.
	EnterClassType(c *ClassTypeContext)

	// EnterInterfaceType is called when entering the interfaceType production.
	EnterInterfaceType(c *InterfaceTypeContext)

	// EnterTypeVariable is called when entering the typeVariable production.
	EnterTypeVariable(c *TypeVariableContext)

	// EnterArrayType is called when entering the arrayType production.
	EnterArrayType(c *ArrayTypeContext)

	// EnterDims is called when entering the dims production.
	EnterDims(c *DimsContext)

	// EnterTypeParameter is called when entering the typeParameter production.
	EnterTypeParameter(c *TypeParameterContext)

	// EnterTypeParameterModifier is called when entering the typeParameterModifier production.
	EnterTypeParameterModifier(c *TypeParameterModifierContext)

	// EnterTypeBound is called when entering the typeBound production.
	EnterTypeBound(c *TypeBoundContext)

	// EnterAdditionalBound is called when entering the additionalBound production.
	EnterAdditionalBound(c *AdditionalBoundContext)

	// EnterTypeArguments is called when entering the typeArguments production.
	EnterTypeArguments(c *TypeArgumentsContext)

	// EnterTypeArgumentList is called when entering the typeArgumentList production.
	EnterTypeArgumentList(c *TypeArgumentListContext)

	// EnterTypeArgument is called when entering the typeArgument production.
	EnterTypeArgument(c *TypeArgumentContext)

	// EnterWildcard is called when entering the wildcard production.
	EnterWildcard(c *WildcardContext)

	// EnterWildcardBounds is called when entering the wildcardBounds production.
	EnterWildcardBounds(c *WildcardBoundsContext)

	// EnterModuleName is called when entering the moduleName production.
	EnterModuleName(c *ModuleNameContext)

	// EnterPackageName is called when entering the packageName production.
	EnterPackageName(c *PackageNameContext)

	// EnterTypeName is called when entering the typeName production.
	EnterTypeName(c *TypeNameContext)

	// EnterPackageOrTypeName is called when entering the packageOrTypeName production.
	EnterPackageOrTypeName(c *PackageOrTypeNameContext)

	// EnterExpressionName is called when entering the expressionName production.
	EnterExpressionName(c *ExpressionNameContext)

	// EnterMethodName is called when entering the methodName production.
	EnterMethodName(c *MethodNameContext)

	// EnterAmbiguousName is called when entering the ambiguousName production.
	EnterAmbiguousName(c *AmbiguousNameContext)

	// EnterCompilationUnit is called when entering the compilationUnit production.
	EnterCompilationUnit(c *CompilationUnitContext)

	// EnterOrdinaryCompilationUnit is called when entering the ordinaryCompilationUnit production.
	EnterOrdinaryCompilationUnit(c *OrdinaryCompilationUnitContext)

	// EnterModularCompilationUnit is called when entering the modularCompilationUnit production.
	EnterModularCompilationUnit(c *ModularCompilationUnitContext)

	// EnterPackageDeclaration is called when entering the packageDeclaration production.
	EnterPackageDeclaration(c *PackageDeclarationContext)

	// EnterPackageModifier is called when entering the packageModifier production.
	EnterPackageModifier(c *PackageModifierContext)

	// EnterImportDeclaration is called when entering the importDeclaration production.
	EnterImportDeclaration(c *ImportDeclarationContext)

	// EnterSingleTypeImportDeclaration is called when entering the singleTypeImportDeclaration production.
	EnterSingleTypeImportDeclaration(c *SingleTypeImportDeclarationContext)

	// EnterTypeImportOnDemandDeclaration is called when entering the typeImportOnDemandDeclaration production.
	EnterTypeImportOnDemandDeclaration(c *TypeImportOnDemandDeclarationContext)

	// EnterSingleStaticImportDeclaration is called when entering the singleStaticImportDeclaration production.
	EnterSingleStaticImportDeclaration(c *SingleStaticImportDeclarationContext)

	// EnterStaticImportOnDemandDeclaration is called when entering the staticImportOnDemandDeclaration production.
	EnterStaticImportOnDemandDeclaration(c *StaticImportOnDemandDeclarationContext)

	// EnterTopLevelClassOrInterfaceDeclaration is called when entering the topLevelClassOrInterfaceDeclaration production.
	EnterTopLevelClassOrInterfaceDeclaration(c *TopLevelClassOrInterfaceDeclarationContext)

	// EnterModuleDeclaration is called when entering the moduleDeclaration production.
	EnterModuleDeclaration(c *ModuleDeclarationContext)

	// EnterModuleDirective is called when entering the moduleDirective production.
	EnterModuleDirective(c *ModuleDirectiveContext)

	// EnterRequiresModifier is called when entering the requiresModifier production.
	EnterRequiresModifier(c *RequiresModifierContext)

	// EnterClassDeclaration is called when entering the classDeclaration production.
	EnterClassDeclaration(c *ClassDeclarationContext)

	// EnterNormalClassDeclaration is called when entering the normalClassDeclaration production.
	EnterNormalClassDeclaration(c *NormalClassDeclarationContext)

	// EnterClassModifier is called when entering the classModifier production.
	EnterClassModifier(c *ClassModifierContext)

	// EnterTypeParameters is called when entering the typeParameters production.
	EnterTypeParameters(c *TypeParametersContext)

	// EnterTypeParameterList is called when entering the typeParameterList production.
	EnterTypeParameterList(c *TypeParameterListContext)

	// EnterClassExtends is called when entering the classExtends production.
	EnterClassExtends(c *ClassExtendsContext)

	// EnterClassImplements is called when entering the classImplements production.
	EnterClassImplements(c *ClassImplementsContext)

	// EnterInterfaceTypeList is called when entering the interfaceTypeList production.
	EnterInterfaceTypeList(c *InterfaceTypeListContext)

	// EnterClassPermits is called when entering the classPermits production.
	EnterClassPermits(c *ClassPermitsContext)

	// EnterClassBody is called when entering the classBody production.
	EnterClassBody(c *ClassBodyContext)

	// EnterClassBodyDeclaration is called when entering the classBodyDeclaration production.
	EnterClassBodyDeclaration(c *ClassBodyDeclarationContext)

	// EnterClassMemberDeclaration is called when entering the classMemberDeclaration production.
	EnterClassMemberDeclaration(c *ClassMemberDeclarationContext)

	// EnterFieldDeclaration is called when entering the fieldDeclaration production.
	EnterFieldDeclaration(c *FieldDeclarationContext)

	// EnterFieldModifier is called when entering the fieldModifier production.
	EnterFieldModifier(c *FieldModifierContext)

	// EnterVariableDeclaratorList is called when entering the variableDeclaratorList production.
	EnterVariableDeclaratorList(c *VariableDeclaratorListContext)

	// EnterVariableDeclarator is called when entering the variableDeclarator production.
	EnterVariableDeclarator(c *VariableDeclaratorContext)

	// EnterVariableDeclaratorId is called when entering the variableDeclaratorId production.
	EnterVariableDeclaratorId(c *VariableDeclaratorIdContext)

	// EnterVariableInitializer is called when entering the variableInitializer production.
	EnterVariableInitializer(c *VariableInitializerContext)

	// EnterUnannType is called when entering the unannType production.
	EnterUnannType(c *UnannTypeContext)

	// EnterUnannPrimitiveType is called when entering the unannPrimitiveType production.
	EnterUnannPrimitiveType(c *UnannPrimitiveTypeContext)

	// EnterUnannReferenceType is called when entering the unannReferenceType production.
	EnterUnannReferenceType(c *UnannReferenceTypeContext)

	// EnterUnannClassOrInterfaceType is called when entering the unannClassOrInterfaceType production.
	EnterUnannClassOrInterfaceType(c *UnannClassOrInterfaceTypeContext)

	// EnterUCOIT is called when entering the uCOIT production.
	EnterUCOIT(c *UCOITContext)

	// EnterUnannClassType is called when entering the unannClassType production.
	EnterUnannClassType(c *UnannClassTypeContext)

	// EnterUnannInterfaceType is called when entering the unannInterfaceType production.
	EnterUnannInterfaceType(c *UnannInterfaceTypeContext)

	// EnterUnannTypeVariable is called when entering the unannTypeVariable production.
	EnterUnannTypeVariable(c *UnannTypeVariableContext)

	// EnterUnannArrayType is called when entering the unannArrayType production.
	EnterUnannArrayType(c *UnannArrayTypeContext)

	// EnterMethodDeclaration is called when entering the methodDeclaration production.
	EnterMethodDeclaration(c *MethodDeclarationContext)

	// EnterMethodModifier is called when entering the methodModifier production.
	EnterMethodModifier(c *MethodModifierContext)

	// EnterMethodHeader is called when entering the methodHeader production.
	EnterMethodHeader(c *MethodHeaderContext)

	// EnterResult is called when entering the result production.
	EnterResult(c *ResultContext)

	// EnterMethodDeclarator is called when entering the methodDeclarator production.
	EnterMethodDeclarator(c *MethodDeclaratorContext)

	// EnterReceiverParameter is called when entering the receiverParameter production.
	EnterReceiverParameter(c *ReceiverParameterContext)

	// EnterFormalParameterList is called when entering the formalParameterList production.
	EnterFormalParameterList(c *FormalParameterListContext)

	// EnterFormalParameter is called when entering the formalParameter production.
	EnterFormalParameter(c *FormalParameterContext)

	// EnterVariableArityParameter is called when entering the variableArityParameter production.
	EnterVariableArityParameter(c *VariableArityParameterContext)

	// EnterVariableModifier is called when entering the variableModifier production.
	EnterVariableModifier(c *VariableModifierContext)

	// EnterThrowsT is called when entering the throwsT production.
	EnterThrowsT(c *ThrowsTContext)

	// EnterExceptionTypeList is called when entering the exceptionTypeList production.
	EnterExceptionTypeList(c *ExceptionTypeListContext)

	// EnterExceptionType is called when entering the exceptionType production.
	EnterExceptionType(c *ExceptionTypeContext)

	// EnterMethodBody is called when entering the methodBody production.
	EnterMethodBody(c *MethodBodyContext)

	// EnterInstanceInitializer is called when entering the instanceInitializer production.
	EnterInstanceInitializer(c *InstanceInitializerContext)

	// EnterStaticInitializer is called when entering the staticInitializer production.
	EnterStaticInitializer(c *StaticInitializerContext)

	// EnterConstructorDeclaration is called when entering the constructorDeclaration production.
	EnterConstructorDeclaration(c *ConstructorDeclarationContext)

	// EnterConstructorModifier is called when entering the constructorModifier production.
	EnterConstructorModifier(c *ConstructorModifierContext)

	// EnterConstructorDeclarator is called when entering the constructorDeclarator production.
	EnterConstructorDeclarator(c *ConstructorDeclaratorContext)

	// EnterSimpleTypeName is called when entering the simpleTypeName production.
	EnterSimpleTypeName(c *SimpleTypeNameContext)

	// EnterConstructorBody is called when entering the constructorBody production.
	EnterConstructorBody(c *ConstructorBodyContext)

	// EnterExplicitConstructorInvocation is called when entering the explicitConstructorInvocation production.
	EnterExplicitConstructorInvocation(c *ExplicitConstructorInvocationContext)

	// EnterEnumDeclaration is called when entering the enumDeclaration production.
	EnterEnumDeclaration(c *EnumDeclarationContext)

	// EnterEnumBody is called when entering the enumBody production.
	EnterEnumBody(c *EnumBodyContext)

	// EnterEnumConstantList is called when entering the enumConstantList production.
	EnterEnumConstantList(c *EnumConstantListContext)

	// EnterEnumConstant is called when entering the enumConstant production.
	EnterEnumConstant(c *EnumConstantContext)

	// EnterEnumConstantModifier is called when entering the enumConstantModifier production.
	EnterEnumConstantModifier(c *EnumConstantModifierContext)

	// EnterEnumBodyDeclarations is called when entering the enumBodyDeclarations production.
	EnterEnumBodyDeclarations(c *EnumBodyDeclarationsContext)

	// EnterRecordDeclaration is called when entering the recordDeclaration production.
	EnterRecordDeclaration(c *RecordDeclarationContext)

	// EnterRecordHeader is called when entering the recordHeader production.
	EnterRecordHeader(c *RecordHeaderContext)

	// EnterRecordComponentList is called when entering the recordComponentList production.
	EnterRecordComponentList(c *RecordComponentListContext)

	// EnterRecordComponent is called when entering the recordComponent production.
	EnterRecordComponent(c *RecordComponentContext)

	// EnterVariableArityRecordComponent is called when entering the variableArityRecordComponent production.
	EnterVariableArityRecordComponent(c *VariableArityRecordComponentContext)

	// EnterRecordComponentModifier is called when entering the recordComponentModifier production.
	EnterRecordComponentModifier(c *RecordComponentModifierContext)

	// EnterRecordBody is called when entering the recordBody production.
	EnterRecordBody(c *RecordBodyContext)

	// EnterRecordBodyDeclaration is called when entering the recordBodyDeclaration production.
	EnterRecordBodyDeclaration(c *RecordBodyDeclarationContext)

	// EnterCompactConstructorDeclaration is called when entering the compactConstructorDeclaration production.
	EnterCompactConstructorDeclaration(c *CompactConstructorDeclarationContext)

	// EnterInterfaceDeclaration is called when entering the interfaceDeclaration production.
	EnterInterfaceDeclaration(c *InterfaceDeclarationContext)

	// EnterNormalInterfaceDeclaration is called when entering the normalInterfaceDeclaration production.
	EnterNormalInterfaceDeclaration(c *NormalInterfaceDeclarationContext)

	// EnterInterfaceModifier is called when entering the interfaceModifier production.
	EnterInterfaceModifier(c *InterfaceModifierContext)

	// EnterInterfaceExtends is called when entering the interfaceExtends production.
	EnterInterfaceExtends(c *InterfaceExtendsContext)

	// EnterInterfacePermits is called when entering the interfacePermits production.
	EnterInterfacePermits(c *InterfacePermitsContext)

	// EnterInterfaceBody is called when entering the interfaceBody production.
	EnterInterfaceBody(c *InterfaceBodyContext)

	// EnterInterfaceMemberDeclaration is called when entering the interfaceMemberDeclaration production.
	EnterInterfaceMemberDeclaration(c *InterfaceMemberDeclarationContext)

	// EnterConstantDeclaration is called when entering the constantDeclaration production.
	EnterConstantDeclaration(c *ConstantDeclarationContext)

	// EnterConstantModifier is called when entering the constantModifier production.
	EnterConstantModifier(c *ConstantModifierContext)

	// EnterInterfaceMethodDeclaration is called when entering the interfaceMethodDeclaration production.
	EnterInterfaceMethodDeclaration(c *InterfaceMethodDeclarationContext)

	// EnterInterfaceMethodModifier is called when entering the interfaceMethodModifier production.
	EnterInterfaceMethodModifier(c *InterfaceMethodModifierContext)

	// EnterAnnotationInterfaceDeclaration is called when entering the annotationInterfaceDeclaration production.
	EnterAnnotationInterfaceDeclaration(c *AnnotationInterfaceDeclarationContext)

	// EnterAnnotationInterfaceBody is called when entering the annotationInterfaceBody production.
	EnterAnnotationInterfaceBody(c *AnnotationInterfaceBodyContext)

	// EnterAnnotationInterfaceMemberDeclaration is called when entering the annotationInterfaceMemberDeclaration production.
	EnterAnnotationInterfaceMemberDeclaration(c *AnnotationInterfaceMemberDeclarationContext)

	// EnterAnnotationInterfaceElementDeclaration is called when entering the annotationInterfaceElementDeclaration production.
	EnterAnnotationInterfaceElementDeclaration(c *AnnotationInterfaceElementDeclarationContext)

	// EnterAnnotationInterfaceElementModifier is called when entering the annotationInterfaceElementModifier production.
	EnterAnnotationInterfaceElementModifier(c *AnnotationInterfaceElementModifierContext)

	// EnterDefaultValue is called when entering the defaultValue production.
	EnterDefaultValue(c *DefaultValueContext)

	// EnterAnnotation is called when entering the annotation production.
	EnterAnnotation(c *AnnotationContext)

	// EnterNormalAnnotation is called when entering the normalAnnotation production.
	EnterNormalAnnotation(c *NormalAnnotationContext)

	// EnterElementValuePairList is called when entering the elementValuePairList production.
	EnterElementValuePairList(c *ElementValuePairListContext)

	// EnterElementValuePair is called when entering the elementValuePair production.
	EnterElementValuePair(c *ElementValuePairContext)

	// EnterElementValue is called when entering the elementValue production.
	EnterElementValue(c *ElementValueContext)

	// EnterElementValueArrayInitializer is called when entering the elementValueArrayInitializer production.
	EnterElementValueArrayInitializer(c *ElementValueArrayInitializerContext)

	// EnterElementValueList is called when entering the elementValueList production.
	EnterElementValueList(c *ElementValueListContext)

	// EnterMarkerAnnotation is called when entering the markerAnnotation production.
	EnterMarkerAnnotation(c *MarkerAnnotationContext)

	// EnterSingleElementAnnotation is called when entering the singleElementAnnotation production.
	EnterSingleElementAnnotation(c *SingleElementAnnotationContext)

	// EnterArrayInitializer is called when entering the arrayInitializer production.
	EnterArrayInitializer(c *ArrayInitializerContext)

	// EnterVariableInitializerList is called when entering the variableInitializerList production.
	EnterVariableInitializerList(c *VariableInitializerListContext)

	// EnterBlock is called when entering the block production.
	EnterBlock(c *BlockContext)

	// EnterBlockStatements is called when entering the blockStatements production.
	EnterBlockStatements(c *BlockStatementsContext)

	// EnterBlockStatement is called when entering the blockStatement production.
	EnterBlockStatement(c *BlockStatementContext)

	// EnterLocalClassOrInterfaceDeclaration is called when entering the localClassOrInterfaceDeclaration production.
	EnterLocalClassOrInterfaceDeclaration(c *LocalClassOrInterfaceDeclarationContext)

	// EnterLocalVariableDeclaration is called when entering the localVariableDeclaration production.
	EnterLocalVariableDeclaration(c *LocalVariableDeclarationContext)

	// EnterLocalVariableType is called when entering the localVariableType production.
	EnterLocalVariableType(c *LocalVariableTypeContext)

	// EnterLocalVariableDeclarationStatement is called when entering the localVariableDeclarationStatement production.
	EnterLocalVariableDeclarationStatement(c *LocalVariableDeclarationStatementContext)

	// EnterStatement is called when entering the statement production.
	EnterStatement(c *StatementContext)

	// EnterStatementNoShortIf is called when entering the statementNoShortIf production.
	EnterStatementNoShortIf(c *StatementNoShortIfContext)

	// EnterStatementWithoutTrailingSubstatement is called when entering the statementWithoutTrailingSubstatement production.
	EnterStatementWithoutTrailingSubstatement(c *StatementWithoutTrailingSubstatementContext)

	// EnterEmptyStatement_ is called when entering the emptyStatement_ production.
	EnterEmptyStatement_(c *EmptyStatement_Context)

	// EnterLabeledStatement is called when entering the labeledStatement production.
	EnterLabeledStatement(c *LabeledStatementContext)

	// EnterLabeledStatementNoShortIf is called when entering the labeledStatementNoShortIf production.
	EnterLabeledStatementNoShortIf(c *LabeledStatementNoShortIfContext)

	// EnterExpressionStatement is called when entering the expressionStatement production.
	EnterExpressionStatement(c *ExpressionStatementContext)

	// EnterStatementExpression is called when entering the statementExpression production.
	EnterStatementExpression(c *StatementExpressionContext)

	// EnterIfThenStatement is called when entering the ifThenStatement production.
	EnterIfThenStatement(c *IfThenStatementContext)

	// EnterIfThenElseStatement is called when entering the ifThenElseStatement production.
	EnterIfThenElseStatement(c *IfThenElseStatementContext)

	// EnterIfThenElseStatementNoShortIf is called when entering the ifThenElseStatementNoShortIf production.
	EnterIfThenElseStatementNoShortIf(c *IfThenElseStatementNoShortIfContext)

	// EnterAssertStatement is called when entering the assertStatement production.
	EnterAssertStatement(c *AssertStatementContext)

	// EnterSwitchStatement is called when entering the switchStatement production.
	EnterSwitchStatement(c *SwitchStatementContext)

	// EnterSwitchBlock is called when entering the switchBlock production.
	EnterSwitchBlock(c *SwitchBlockContext)

	// EnterSwitchRule is called when entering the switchRule production.
	EnterSwitchRule(c *SwitchRuleContext)

	// EnterSwitchBlockStatementGroup is called when entering the switchBlockStatementGroup production.
	EnterSwitchBlockStatementGroup(c *SwitchBlockStatementGroupContext)

	// EnterSwitchLabel is called when entering the switchLabel production.
	EnterSwitchLabel(c *SwitchLabelContext)

	// EnterCaseConstant is called when entering the caseConstant production.
	EnterCaseConstant(c *CaseConstantContext)

	// EnterWhileStatement is called when entering the whileStatement production.
	EnterWhileStatement(c *WhileStatementContext)

	// EnterWhileStatementNoShortIf is called when entering the whileStatementNoShortIf production.
	EnterWhileStatementNoShortIf(c *WhileStatementNoShortIfContext)

	// EnterDoStatement is called when entering the doStatement production.
	EnterDoStatement(c *DoStatementContext)

	// EnterForStatement is called when entering the forStatement production.
	EnterForStatement(c *ForStatementContext)

	// EnterForStatementNoShortIf is called when entering the forStatementNoShortIf production.
	EnterForStatementNoShortIf(c *ForStatementNoShortIfContext)

	// EnterBasicForStatement is called when entering the basicForStatement production.
	EnterBasicForStatement(c *BasicForStatementContext)

	// EnterBasicForStatementNoShortIf is called when entering the basicForStatementNoShortIf production.
	EnterBasicForStatementNoShortIf(c *BasicForStatementNoShortIfContext)

	// EnterForInit is called when entering the forInit production.
	EnterForInit(c *ForInitContext)

	// EnterForUpdate is called when entering the forUpdate production.
	EnterForUpdate(c *ForUpdateContext)

	// EnterStatementExpressionList is called when entering the statementExpressionList production.
	EnterStatementExpressionList(c *StatementExpressionListContext)

	// EnterEnhancedForStatement is called when entering the enhancedForStatement production.
	EnterEnhancedForStatement(c *EnhancedForStatementContext)

	// EnterEnhancedForStatementNoShortIf is called when entering the enhancedForStatementNoShortIf production.
	EnterEnhancedForStatementNoShortIf(c *EnhancedForStatementNoShortIfContext)

	// EnterBreakStatement is called when entering the breakStatement production.
	EnterBreakStatement(c *BreakStatementContext)

	// EnterContinueStatement is called when entering the continueStatement production.
	EnterContinueStatement(c *ContinueStatementContext)

	// EnterReturnStatement is called when entering the returnStatement production.
	EnterReturnStatement(c *ReturnStatementContext)

	// EnterThrowStatement is called when entering the throwStatement production.
	EnterThrowStatement(c *ThrowStatementContext)

	// EnterSynchronizedStatement is called when entering the synchronizedStatement production.
	EnterSynchronizedStatement(c *SynchronizedStatementContext)

	// EnterTryStatement is called when entering the tryStatement production.
	EnterTryStatement(c *TryStatementContext)

	// EnterCatches is called when entering the catches production.
	EnterCatches(c *CatchesContext)

	// EnterCatchClause is called when entering the catchClause production.
	EnterCatchClause(c *CatchClauseContext)

	// EnterCatchFormalParameter is called when entering the catchFormalParameter production.
	EnterCatchFormalParameter(c *CatchFormalParameterContext)

	// EnterCatchType is called when entering the catchType production.
	EnterCatchType(c *CatchTypeContext)

	// EnterFinallyBlock is called when entering the finallyBlock production.
	EnterFinallyBlock(c *FinallyBlockContext)

	// EnterTryWithResourcesStatement is called when entering the tryWithResourcesStatement production.
	EnterTryWithResourcesStatement(c *TryWithResourcesStatementContext)

	// EnterResourceSpecification is called when entering the resourceSpecification production.
	EnterResourceSpecification(c *ResourceSpecificationContext)

	// EnterResourceList is called when entering the resourceList production.
	EnterResourceList(c *ResourceListContext)

	// EnterResource is called when entering the resource production.
	EnterResource(c *ResourceContext)

	// EnterVariableAccess is called when entering the variableAccess production.
	EnterVariableAccess(c *VariableAccessContext)

	// EnterYieldStatement is called when entering the yieldStatement production.
	EnterYieldStatement(c *YieldStatementContext)

	// EnterPattern is called when entering the pattern production.
	EnterPattern(c *PatternContext)

	// EnterTypePattern is called when entering the typePattern production.
	EnterTypePattern(c *TypePatternContext)

	// EnterExpression is called when entering the expression production.
	EnterExpression(c *ExpressionContext)

	// EnterPrimary is called when entering the primary production.
	EnterPrimary(c *PrimaryContext)

	// EnterPrimaryNoNewArray is called when entering the primaryNoNewArray production.
	EnterPrimaryNoNewArray(c *PrimaryNoNewArrayContext)

	// EnterPNNA is called when entering the pNNA production.
	EnterPNNA(c *PNNAContext)

	// EnterClassLiteral is called when entering the classLiteral production.
	EnterClassLiteral(c *ClassLiteralContext)

	// EnterClassInstanceCreationExpression is called when entering the classInstanceCreationExpression production.
	EnterClassInstanceCreationExpression(c *ClassInstanceCreationExpressionContext)

	// EnterUnqualifiedClassInstanceCreationExpression is called when entering the unqualifiedClassInstanceCreationExpression production.
	EnterUnqualifiedClassInstanceCreationExpression(c *UnqualifiedClassInstanceCreationExpressionContext)

	// EnterClassOrInterfaceTypeToInstantiate is called when entering the classOrInterfaceTypeToInstantiate production.
	EnterClassOrInterfaceTypeToInstantiate(c *ClassOrInterfaceTypeToInstantiateContext)

	// EnterTypeArgumentsOrDiamond is called when entering the typeArgumentsOrDiamond production.
	EnterTypeArgumentsOrDiamond(c *TypeArgumentsOrDiamondContext)

	// EnterArrayCreationExpression is called when entering the arrayCreationExpression production.
	EnterArrayCreationExpression(c *ArrayCreationExpressionContext)

	// EnterArrayCreationExpressionWithoutInitializer is called when entering the arrayCreationExpressionWithoutInitializer production.
	EnterArrayCreationExpressionWithoutInitializer(c *ArrayCreationExpressionWithoutInitializerContext)

	// EnterArrayCreationExpressionWithInitializer is called when entering the arrayCreationExpressionWithInitializer production.
	EnterArrayCreationExpressionWithInitializer(c *ArrayCreationExpressionWithInitializerContext)

	// EnterDimExprs is called when entering the dimExprs production.
	EnterDimExprs(c *DimExprsContext)

	// EnterDimExpr is called when entering the dimExpr production.
	EnterDimExpr(c *DimExprContext)

	// EnterArrayAccess is called when entering the arrayAccess production.
	EnterArrayAccess(c *ArrayAccessContext)

	// EnterFieldAccess is called when entering the fieldAccess production.
	EnterFieldAccess(c *FieldAccessContext)

	// EnterMethodInvocation is called when entering the methodInvocation production.
	EnterMethodInvocation(c *MethodInvocationContext)

	// EnterArgumentList is called when entering the argumentList production.
	EnterArgumentList(c *ArgumentListContext)

	// EnterMethodReference is called when entering the methodReference production.
	EnterMethodReference(c *MethodReferenceContext)

	// EnterPostfixExpression is called when entering the postfixExpression production.
	EnterPostfixExpression(c *PostfixExpressionContext)

	// EnterPfE is called when entering the pfE production.
	EnterPfE(c *PfEContext)

	// EnterPostIncrementExpression is called when entering the postIncrementExpression production.
	EnterPostIncrementExpression(c *PostIncrementExpressionContext)

	// EnterPostDecrementExpression is called when entering the postDecrementExpression production.
	EnterPostDecrementExpression(c *PostDecrementExpressionContext)

	// EnterUnaryExpression is called when entering the unaryExpression production.
	EnterUnaryExpression(c *UnaryExpressionContext)

	// EnterPreIncrementExpression is called when entering the preIncrementExpression production.
	EnterPreIncrementExpression(c *PreIncrementExpressionContext)

	// EnterPreDecrementExpression is called when entering the preDecrementExpression production.
	EnterPreDecrementExpression(c *PreDecrementExpressionContext)

	// EnterUnaryExpressionNotPlusMinus is called when entering the unaryExpressionNotPlusMinus production.
	EnterUnaryExpressionNotPlusMinus(c *UnaryExpressionNotPlusMinusContext)

	// EnterCastExpression is called when entering the castExpression production.
	EnterCastExpression(c *CastExpressionContext)

	// EnterMultiplicativeExpression is called when entering the multiplicativeExpression production.
	EnterMultiplicativeExpression(c *MultiplicativeExpressionContext)

	// EnterAdditiveExpression is called when entering the additiveExpression production.
	EnterAdditiveExpression(c *AdditiveExpressionContext)

	// EnterShiftExpression is called when entering the shiftExpression production.
	EnterShiftExpression(c *ShiftExpressionContext)

	// EnterRelationalExpression is called when entering the relationalExpression production.
	EnterRelationalExpression(c *RelationalExpressionContext)

	// EnterEqualityExpression is called when entering the equalityExpression production.
	EnterEqualityExpression(c *EqualityExpressionContext)

	// EnterAndExpression is called when entering the andExpression production.
	EnterAndExpression(c *AndExpressionContext)

	// EnterExclusiveOrExpression is called when entering the exclusiveOrExpression production.
	EnterExclusiveOrExpression(c *ExclusiveOrExpressionContext)

	// EnterInclusiveOrExpression is called when entering the inclusiveOrExpression production.
	EnterInclusiveOrExpression(c *InclusiveOrExpressionContext)

	// EnterConditionalAndExpression is called when entering the conditionalAndExpression production.
	EnterConditionalAndExpression(c *ConditionalAndExpressionContext)

	// EnterConditionalOrExpression is called when entering the conditionalOrExpression production.
	EnterConditionalOrExpression(c *ConditionalOrExpressionContext)

	// EnterConditionalExpression is called when entering the conditionalExpression production.
	EnterConditionalExpression(c *ConditionalExpressionContext)

	// EnterAssignmentExpression is called when entering the assignmentExpression production.
	EnterAssignmentExpression(c *AssignmentExpressionContext)

	// EnterAssignment is called when entering the assignment production.
	EnterAssignment(c *AssignmentContext)

	// EnterLeftHandSide is called when entering the leftHandSide production.
	EnterLeftHandSide(c *LeftHandSideContext)

	// EnterAssignmentOperator is called when entering the assignmentOperator production.
	EnterAssignmentOperator(c *AssignmentOperatorContext)

	// EnterLambdaExpression is called when entering the lambdaExpression production.
	EnterLambdaExpression(c *LambdaExpressionContext)

	// EnterLambdaParameters is called when entering the lambdaParameters production.
	EnterLambdaParameters(c *LambdaParametersContext)

	// EnterLambdaParameterList is called when entering the lambdaParameterList production.
	EnterLambdaParameterList(c *LambdaParameterListContext)

	// EnterLambdaParameter is called when entering the lambdaParameter production.
	EnterLambdaParameter(c *LambdaParameterContext)

	// EnterLambdaParameterType is called when entering the lambdaParameterType production.
	EnterLambdaParameterType(c *LambdaParameterTypeContext)

	// EnterLambdaBody is called when entering the lambdaBody production.
	EnterLambdaBody(c *LambdaBodyContext)

	// EnterSwitchExpression is called when entering the switchExpression production.
	EnterSwitchExpression(c *SwitchExpressionContext)

	// EnterConstantExpression is called when entering the constantExpression production.
	EnterConstantExpression(c *ConstantExpressionContext)

	// ExitStart_ is called when exiting the start_ production.
	ExitStart_(c *Start_Context)

	// ExitIdentifier is called when exiting the identifier production.
	ExitIdentifier(c *IdentifierContext)

	// ExitTypeIdentifier is called when exiting the typeIdentifier production.
	ExitTypeIdentifier(c *TypeIdentifierContext)

	// ExitUnqualifiedMethodIdentifier is called when exiting the unqualifiedMethodIdentifier production.
	ExitUnqualifiedMethodIdentifier(c *UnqualifiedMethodIdentifierContext)

	// ExitContextualKeyword is called when exiting the contextualKeyword production.
	ExitContextualKeyword(c *ContextualKeywordContext)

	// ExitContextualKeywordMinusForTypeIdentifier is called when exiting the contextualKeywordMinusForTypeIdentifier production.
	ExitContextualKeywordMinusForTypeIdentifier(c *ContextualKeywordMinusForTypeIdentifierContext)

	// ExitContextualKeywordMinusForUnqualifiedMethodIdentifier is called when exiting the contextualKeywordMinusForUnqualifiedMethodIdentifier production.
	ExitContextualKeywordMinusForUnqualifiedMethodIdentifier(c *ContextualKeywordMinusForUnqualifiedMethodIdentifierContext)

	// ExitLiteral is called when exiting the literal production.
	ExitLiteral(c *LiteralContext)

	// ExitPrimitiveType is called when exiting the primitiveType production.
	ExitPrimitiveType(c *PrimitiveTypeContext)

	// ExitNumericType is called when exiting the numericType production.
	ExitNumericType(c *NumericTypeContext)

	// ExitIntegralType is called when exiting the integralType production.
	ExitIntegralType(c *IntegralTypeContext)

	// ExitFloatingPointType is called when exiting the floatingPointType production.
	ExitFloatingPointType(c *FloatingPointTypeContext)

	// ExitReferenceType is called when exiting the referenceType production.
	ExitReferenceType(c *ReferenceTypeContext)

	// ExitCoit is called when exiting the coit production.
	ExitCoit(c *CoitContext)

	// ExitClassOrInterfaceType is called when exiting the classOrInterfaceType production.
	ExitClassOrInterfaceType(c *ClassOrInterfaceTypeContext)

	// ExitClassType is called when exiting the classType production.
	ExitClassType(c *ClassTypeContext)

	// ExitInterfaceType is called when exiting the interfaceType production.
	ExitInterfaceType(c *InterfaceTypeContext)

	// ExitTypeVariable is called when exiting the typeVariable production.
	ExitTypeVariable(c *TypeVariableContext)

	// ExitArrayType is called when exiting the arrayType production.
	ExitArrayType(c *ArrayTypeContext)

	// ExitDims is called when exiting the dims production.
	ExitDims(c *DimsContext)

	// ExitTypeParameter is called when exiting the typeParameter production.
	ExitTypeParameter(c *TypeParameterContext)

	// ExitTypeParameterModifier is called when exiting the typeParameterModifier production.
	ExitTypeParameterModifier(c *TypeParameterModifierContext)

	// ExitTypeBound is called when exiting the typeBound production.
	ExitTypeBound(c *TypeBoundContext)

	// ExitAdditionalBound is called when exiting the additionalBound production.
	ExitAdditionalBound(c *AdditionalBoundContext)

	// ExitTypeArguments is called when exiting the typeArguments production.
	ExitTypeArguments(c *TypeArgumentsContext)

	// ExitTypeArgumentList is called when exiting the typeArgumentList production.
	ExitTypeArgumentList(c *TypeArgumentListContext)

	// ExitTypeArgument is called when exiting the typeArgument production.
	ExitTypeArgument(c *TypeArgumentContext)

	// ExitWildcard is called when exiting the wildcard production.
	ExitWildcard(c *WildcardContext)

	// ExitWildcardBounds is called when exiting the wildcardBounds production.
	ExitWildcardBounds(c *WildcardBoundsContext)

	// ExitModuleName is called when exiting the moduleName production.
	ExitModuleName(c *ModuleNameContext)

	// ExitPackageName is called when exiting the packageName production.
	ExitPackageName(c *PackageNameContext)

	// ExitTypeName is called when exiting the typeName production.
	ExitTypeName(c *TypeNameContext)

	// ExitPackageOrTypeName is called when exiting the packageOrTypeName production.
	ExitPackageOrTypeName(c *PackageOrTypeNameContext)

	// ExitExpressionName is called when exiting the expressionName production.
	ExitExpressionName(c *ExpressionNameContext)

	// ExitMethodName is called when exiting the methodName production.
	ExitMethodName(c *MethodNameContext)

	// ExitAmbiguousName is called when exiting the ambiguousName production.
	ExitAmbiguousName(c *AmbiguousNameContext)

	// ExitCompilationUnit is called when exiting the compilationUnit production.
	ExitCompilationUnit(c *CompilationUnitContext)

	// ExitOrdinaryCompilationUnit is called when exiting the ordinaryCompilationUnit production.
	ExitOrdinaryCompilationUnit(c *OrdinaryCompilationUnitContext)

	// ExitModularCompilationUnit is called when exiting the modularCompilationUnit production.
	ExitModularCompilationUnit(c *ModularCompilationUnitContext)

	// ExitPackageDeclaration is called when exiting the packageDeclaration production.
	ExitPackageDeclaration(c *PackageDeclarationContext)

	// ExitPackageModifier is called when exiting the packageModifier production.
	ExitPackageModifier(c *PackageModifierContext)

	// ExitImportDeclaration is called when exiting the importDeclaration production.
	ExitImportDeclaration(c *ImportDeclarationContext)

	// ExitSingleTypeImportDeclaration is called when exiting the singleTypeImportDeclaration production.
	ExitSingleTypeImportDeclaration(c *SingleTypeImportDeclarationContext)

	// ExitTypeImportOnDemandDeclaration is called when exiting the typeImportOnDemandDeclaration production.
	ExitTypeImportOnDemandDeclaration(c *TypeImportOnDemandDeclarationContext)

	// ExitSingleStaticImportDeclaration is called when exiting the singleStaticImportDeclaration production.
	ExitSingleStaticImportDeclaration(c *SingleStaticImportDeclarationContext)

	// ExitStaticImportOnDemandDeclaration is called when exiting the staticImportOnDemandDeclaration production.
	ExitStaticImportOnDemandDeclaration(c *StaticImportOnDemandDeclarationContext)

	// ExitTopLevelClassOrInterfaceDeclaration is called when exiting the topLevelClassOrInterfaceDeclaration production.
	ExitTopLevelClassOrInterfaceDeclaration(c *TopLevelClassOrInterfaceDeclarationContext)

	// ExitModuleDeclaration is called when exiting the moduleDeclaration production.
	ExitModuleDeclaration(c *ModuleDeclarationContext)

	// ExitModuleDirective is called when exiting the moduleDirective production.
	ExitModuleDirective(c *ModuleDirectiveContext)

	// ExitRequiresModifier is called when exiting the requiresModifier production.
	ExitRequiresModifier(c *RequiresModifierContext)

	// ExitClassDeclaration is called when exiting the classDeclaration production.
	ExitClassDeclaration(c *ClassDeclarationContext)

	// ExitNormalClassDeclaration is called when exiting the normalClassDeclaration production.
	ExitNormalClassDeclaration(c *NormalClassDeclarationContext)

	// ExitClassModifier is called when exiting the classModifier production.
	ExitClassModifier(c *ClassModifierContext)

	// ExitTypeParameters is called when exiting the typeParameters production.
	ExitTypeParameters(c *TypeParametersContext)

	// ExitTypeParameterList is called when exiting the typeParameterList production.
	ExitTypeParameterList(c *TypeParameterListContext)

	// ExitClassExtends is called when exiting the classExtends production.
	ExitClassExtends(c *ClassExtendsContext)

	// ExitClassImplements is called when exiting the classImplements production.
	ExitClassImplements(c *ClassImplementsContext)

	// ExitInterfaceTypeList is called when exiting the interfaceTypeList production.
	ExitInterfaceTypeList(c *InterfaceTypeListContext)

	// ExitClassPermits is called when exiting the classPermits production.
	ExitClassPermits(c *ClassPermitsContext)

	// ExitClassBody is called when exiting the classBody production.
	ExitClassBody(c *ClassBodyContext)

	// ExitClassBodyDeclaration is called when exiting the classBodyDeclaration production.
	ExitClassBodyDeclaration(c *ClassBodyDeclarationContext)

	// ExitClassMemberDeclaration is called when exiting the classMemberDeclaration production.
	ExitClassMemberDeclaration(c *ClassMemberDeclarationContext)

	// ExitFieldDeclaration is called when exiting the fieldDeclaration production.
	ExitFieldDeclaration(c *FieldDeclarationContext)

	// ExitFieldModifier is called when exiting the fieldModifier production.
	ExitFieldModifier(c *FieldModifierContext)

	// ExitVariableDeclaratorList is called when exiting the variableDeclaratorList production.
	ExitVariableDeclaratorList(c *VariableDeclaratorListContext)

	// ExitVariableDeclarator is called when exiting the variableDeclarator production.
	ExitVariableDeclarator(c *VariableDeclaratorContext)

	// ExitVariableDeclaratorId is called when exiting the variableDeclaratorId production.
	ExitVariableDeclaratorId(c *VariableDeclaratorIdContext)

	// ExitVariableInitializer is called when exiting the variableInitializer production.
	ExitVariableInitializer(c *VariableInitializerContext)

	// ExitUnannType is called when exiting the unannType production.
	ExitUnannType(c *UnannTypeContext)

	// ExitUnannPrimitiveType is called when exiting the unannPrimitiveType production.
	ExitUnannPrimitiveType(c *UnannPrimitiveTypeContext)

	// ExitUnannReferenceType is called when exiting the unannReferenceType production.
	ExitUnannReferenceType(c *UnannReferenceTypeContext)

	// ExitUnannClassOrInterfaceType is called when exiting the unannClassOrInterfaceType production.
	ExitUnannClassOrInterfaceType(c *UnannClassOrInterfaceTypeContext)

	// ExitUCOIT is called when exiting the uCOIT production.
	ExitUCOIT(c *UCOITContext)

	// ExitUnannClassType is called when exiting the unannClassType production.
	ExitUnannClassType(c *UnannClassTypeContext)

	// ExitUnannInterfaceType is called when exiting the unannInterfaceType production.
	ExitUnannInterfaceType(c *UnannInterfaceTypeContext)

	// ExitUnannTypeVariable is called when exiting the unannTypeVariable production.
	ExitUnannTypeVariable(c *UnannTypeVariableContext)

	// ExitUnannArrayType is called when exiting the unannArrayType production.
	ExitUnannArrayType(c *UnannArrayTypeContext)

	// ExitMethodDeclaration is called when exiting the methodDeclaration production.
	ExitMethodDeclaration(c *MethodDeclarationContext)

	// ExitMethodModifier is called when exiting the methodModifier production.
	ExitMethodModifier(c *MethodModifierContext)

	// ExitMethodHeader is called when exiting the methodHeader production.
	ExitMethodHeader(c *MethodHeaderContext)

	// ExitResult is called when exiting the result production.
	ExitResult(c *ResultContext)

	// ExitMethodDeclarator is called when exiting the methodDeclarator production.
	ExitMethodDeclarator(c *MethodDeclaratorContext)

	// ExitReceiverParameter is called when exiting the receiverParameter production.
	ExitReceiverParameter(c *ReceiverParameterContext)

	// ExitFormalParameterList is called when exiting the formalParameterList production.
	ExitFormalParameterList(c *FormalParameterListContext)

	// ExitFormalParameter is called when exiting the formalParameter production.
	ExitFormalParameter(c *FormalParameterContext)

	// ExitVariableArityParameter is called when exiting the variableArityParameter production.
	ExitVariableArityParameter(c *VariableArityParameterContext)

	// ExitVariableModifier is called when exiting the variableModifier production.
	ExitVariableModifier(c *VariableModifierContext)

	// ExitThrowsT is called when exiting the throwsT production.
	ExitThrowsT(c *ThrowsTContext)

	// ExitExceptionTypeList is called when exiting the exceptionTypeList production.
	ExitExceptionTypeList(c *ExceptionTypeListContext)

	// ExitExceptionType is called when exiting the exceptionType production.
	ExitExceptionType(c *ExceptionTypeContext)

	// ExitMethodBody is called when exiting the methodBody production.
	ExitMethodBody(c *MethodBodyContext)

	// ExitInstanceInitializer is called when exiting the instanceInitializer production.
	ExitInstanceInitializer(c *InstanceInitializerContext)

	// ExitStaticInitializer is called when exiting the staticInitializer production.
	ExitStaticInitializer(c *StaticInitializerContext)

	// ExitConstructorDeclaration is called when exiting the constructorDeclaration production.
	ExitConstructorDeclaration(c *ConstructorDeclarationContext)

	// ExitConstructorModifier is called when exiting the constructorModifier production.
	ExitConstructorModifier(c *ConstructorModifierContext)

	// ExitConstructorDeclarator is called when exiting the constructorDeclarator production.
	ExitConstructorDeclarator(c *ConstructorDeclaratorContext)

	// ExitSimpleTypeName is called when exiting the simpleTypeName production.
	ExitSimpleTypeName(c *SimpleTypeNameContext)

	// ExitConstructorBody is called when exiting the constructorBody production.
	ExitConstructorBody(c *ConstructorBodyContext)

	// ExitExplicitConstructorInvocation is called when exiting the explicitConstructorInvocation production.
	ExitExplicitConstructorInvocation(c *ExplicitConstructorInvocationContext)

	// ExitEnumDeclaration is called when exiting the enumDeclaration production.
	ExitEnumDeclaration(c *EnumDeclarationContext)

	// ExitEnumBody is called when exiting the enumBody production.
	ExitEnumBody(c *EnumBodyContext)

	// ExitEnumConstantList is called when exiting the enumConstantList production.
	ExitEnumConstantList(c *EnumConstantListContext)

	// ExitEnumConstant is called when exiting the enumConstant production.
	ExitEnumConstant(c *EnumConstantContext)

	// ExitEnumConstantModifier is called when exiting the enumConstantModifier production.
	ExitEnumConstantModifier(c *EnumConstantModifierContext)

	// ExitEnumBodyDeclarations is called when exiting the enumBodyDeclarations production.
	ExitEnumBodyDeclarations(c *EnumBodyDeclarationsContext)

	// ExitRecordDeclaration is called when exiting the recordDeclaration production.
	ExitRecordDeclaration(c *RecordDeclarationContext)

	// ExitRecordHeader is called when exiting the recordHeader production.
	ExitRecordHeader(c *RecordHeaderContext)

	// ExitRecordComponentList is called when exiting the recordComponentList production.
	ExitRecordComponentList(c *RecordComponentListContext)

	// ExitRecordComponent is called when exiting the recordComponent production.
	ExitRecordComponent(c *RecordComponentContext)

	// ExitVariableArityRecordComponent is called when exiting the variableArityRecordComponent production.
	ExitVariableArityRecordComponent(c *VariableArityRecordComponentContext)

	// ExitRecordComponentModifier is called when exiting the recordComponentModifier production.
	ExitRecordComponentModifier(c *RecordComponentModifierContext)

	// ExitRecordBody is called when exiting the recordBody production.
	ExitRecordBody(c *RecordBodyContext)

	// ExitRecordBodyDeclaration is called when exiting the recordBodyDeclaration production.
	ExitRecordBodyDeclaration(c *RecordBodyDeclarationContext)

	// ExitCompactConstructorDeclaration is called when exiting the compactConstructorDeclaration production.
	ExitCompactConstructorDeclaration(c *CompactConstructorDeclarationContext)

	// ExitInterfaceDeclaration is called when exiting the interfaceDeclaration production.
	ExitInterfaceDeclaration(c *InterfaceDeclarationContext)

	// ExitNormalInterfaceDeclaration is called when exiting the normalInterfaceDeclaration production.
	ExitNormalInterfaceDeclaration(c *NormalInterfaceDeclarationContext)

	// ExitInterfaceModifier is called when exiting the interfaceModifier production.
	ExitInterfaceModifier(c *InterfaceModifierContext)

	// ExitInterfaceExtends is called when exiting the interfaceExtends production.
	ExitInterfaceExtends(c *InterfaceExtendsContext)

	// ExitInterfacePermits is called when exiting the interfacePermits production.
	ExitInterfacePermits(c *InterfacePermitsContext)

	// ExitInterfaceBody is called when exiting the interfaceBody production.
	ExitInterfaceBody(c *InterfaceBodyContext)

	// ExitInterfaceMemberDeclaration is called when exiting the interfaceMemberDeclaration production.
	ExitInterfaceMemberDeclaration(c *InterfaceMemberDeclarationContext)

	// ExitConstantDeclaration is called when exiting the constantDeclaration production.
	ExitConstantDeclaration(c *ConstantDeclarationContext)

	// ExitConstantModifier is called when exiting the constantModifier production.
	ExitConstantModifier(c *ConstantModifierContext)

	// ExitInterfaceMethodDeclaration is called when exiting the interfaceMethodDeclaration production.
	ExitInterfaceMethodDeclaration(c *InterfaceMethodDeclarationContext)

	// ExitInterfaceMethodModifier is called when exiting the interfaceMethodModifier production.
	ExitInterfaceMethodModifier(c *InterfaceMethodModifierContext)

	// ExitAnnotationInterfaceDeclaration is called when exiting the annotationInterfaceDeclaration production.
	ExitAnnotationInterfaceDeclaration(c *AnnotationInterfaceDeclarationContext)

	// ExitAnnotationInterfaceBody is called when exiting the annotationInterfaceBody production.
	ExitAnnotationInterfaceBody(c *AnnotationInterfaceBodyContext)

	// ExitAnnotationInterfaceMemberDeclaration is called when exiting the annotationInterfaceMemberDeclaration production.
	ExitAnnotationInterfaceMemberDeclaration(c *AnnotationInterfaceMemberDeclarationContext)

	// ExitAnnotationInterfaceElementDeclaration is called when exiting the annotationInterfaceElementDeclaration production.
	ExitAnnotationInterfaceElementDeclaration(c *AnnotationInterfaceElementDeclarationContext)

	// ExitAnnotationInterfaceElementModifier is called when exiting the annotationInterfaceElementModifier production.
	ExitAnnotationInterfaceElementModifier(c *AnnotationInterfaceElementModifierContext)

	// ExitDefaultValue is called when exiting the defaultValue production.
	ExitDefaultValue(c *DefaultValueContext)

	// ExitAnnotation is called when exiting the annotation production.
	ExitAnnotation(c *AnnotationContext)

	// ExitNormalAnnotation is called when exiting the normalAnnotation production.
	ExitNormalAnnotation(c *NormalAnnotationContext)

	// ExitElementValuePairList is called when exiting the elementValuePairList production.
	ExitElementValuePairList(c *ElementValuePairListContext)

	// ExitElementValuePair is called when exiting the elementValuePair production.
	ExitElementValuePair(c *ElementValuePairContext)

	// ExitElementValue is called when exiting the elementValue production.
	ExitElementValue(c *ElementValueContext)

	// ExitElementValueArrayInitializer is called when exiting the elementValueArrayInitializer production.
	ExitElementValueArrayInitializer(c *ElementValueArrayInitializerContext)

	// ExitElementValueList is called when exiting the elementValueList production.
	ExitElementValueList(c *ElementValueListContext)

	// ExitMarkerAnnotation is called when exiting the markerAnnotation production.
	ExitMarkerAnnotation(c *MarkerAnnotationContext)

	// ExitSingleElementAnnotation is called when exiting the singleElementAnnotation production.
	ExitSingleElementAnnotation(c *SingleElementAnnotationContext)

	// ExitArrayInitializer is called when exiting the arrayInitializer production.
	ExitArrayInitializer(c *ArrayInitializerContext)

	// ExitVariableInitializerList is called when exiting the variableInitializerList production.
	ExitVariableInitializerList(c *VariableInitializerListContext)

	// ExitBlock is called when exiting the block production.
	ExitBlock(c *BlockContext)

	// ExitBlockStatements is called when exiting the blockStatements production.
	ExitBlockStatements(c *BlockStatementsContext)

	// ExitBlockStatement is called when exiting the blockStatement production.
	ExitBlockStatement(c *BlockStatementContext)

	// ExitLocalClassOrInterfaceDeclaration is called when exiting the localClassOrInterfaceDeclaration production.
	ExitLocalClassOrInterfaceDeclaration(c *LocalClassOrInterfaceDeclarationContext)

	// ExitLocalVariableDeclaration is called when exiting the localVariableDeclaration production.
	ExitLocalVariableDeclaration(c *LocalVariableDeclarationContext)

	// ExitLocalVariableType is called when exiting the localVariableType production.
	ExitLocalVariableType(c *LocalVariableTypeContext)

	// ExitLocalVariableDeclarationStatement is called when exiting the localVariableDeclarationStatement production.
	ExitLocalVariableDeclarationStatement(c *LocalVariableDeclarationStatementContext)

	// ExitStatement is called when exiting the statement production.
	ExitStatement(c *StatementContext)

	// ExitStatementNoShortIf is called when exiting the statementNoShortIf production.
	ExitStatementNoShortIf(c *StatementNoShortIfContext)

	// ExitStatementWithoutTrailingSubstatement is called when exiting the statementWithoutTrailingSubstatement production.
	ExitStatementWithoutTrailingSubstatement(c *StatementWithoutTrailingSubstatementContext)

	// ExitEmptyStatement_ is called when exiting the emptyStatement_ production.
	ExitEmptyStatement_(c *EmptyStatement_Context)

	// ExitLabeledStatement is called when exiting the labeledStatement production.
	ExitLabeledStatement(c *LabeledStatementContext)

	// ExitLabeledStatementNoShortIf is called when exiting the labeledStatementNoShortIf production.
	ExitLabeledStatementNoShortIf(c *LabeledStatementNoShortIfContext)

	// ExitExpressionStatement is called when exiting the expressionStatement production.
	ExitExpressionStatement(c *ExpressionStatementContext)

	// ExitStatementExpression is called when exiting the statementExpression production.
	ExitStatementExpression(c *StatementExpressionContext)

	// ExitIfThenStatement is called when exiting the ifThenStatement production.
	ExitIfThenStatement(c *IfThenStatementContext)

	// ExitIfThenElseStatement is called when exiting the ifThenElseStatement production.
	ExitIfThenElseStatement(c *IfThenElseStatementContext)

	// ExitIfThenElseStatementNoShortIf is called when exiting the ifThenElseStatementNoShortIf production.
	ExitIfThenElseStatementNoShortIf(c *IfThenElseStatementNoShortIfContext)

	// ExitAssertStatement is called when exiting the assertStatement production.
	ExitAssertStatement(c *AssertStatementContext)

	// ExitSwitchStatement is called when exiting the switchStatement production.
	ExitSwitchStatement(c *SwitchStatementContext)

	// ExitSwitchBlock is called when exiting the switchBlock production.
	ExitSwitchBlock(c *SwitchBlockContext)

	// ExitSwitchRule is called when exiting the switchRule production.
	ExitSwitchRule(c *SwitchRuleContext)

	// ExitSwitchBlockStatementGroup is called when exiting the switchBlockStatementGroup production.
	ExitSwitchBlockStatementGroup(c *SwitchBlockStatementGroupContext)

	// ExitSwitchLabel is called when exiting the switchLabel production.
	ExitSwitchLabel(c *SwitchLabelContext)

	// ExitCaseConstant is called when exiting the caseConstant production.
	ExitCaseConstant(c *CaseConstantContext)

	// ExitWhileStatement is called when exiting the whileStatement production.
	ExitWhileStatement(c *WhileStatementContext)

	// ExitWhileStatementNoShortIf is called when exiting the whileStatementNoShortIf production.
	ExitWhileStatementNoShortIf(c *WhileStatementNoShortIfContext)

	// ExitDoStatement is called when exiting the doStatement production.
	ExitDoStatement(c *DoStatementContext)

	// ExitForStatement is called when exiting the forStatement production.
	ExitForStatement(c *ForStatementContext)

	// ExitForStatementNoShortIf is called when exiting the forStatementNoShortIf production.
	ExitForStatementNoShortIf(c *ForStatementNoShortIfContext)

	// ExitBasicForStatement is called when exiting the basicForStatement production.
	ExitBasicForStatement(c *BasicForStatementContext)

	// ExitBasicForStatementNoShortIf is called when exiting the basicForStatementNoShortIf production.
	ExitBasicForStatementNoShortIf(c *BasicForStatementNoShortIfContext)

	// ExitForInit is called when exiting the forInit production.
	ExitForInit(c *ForInitContext)

	// ExitForUpdate is called when exiting the forUpdate production.
	ExitForUpdate(c *ForUpdateContext)

	// ExitStatementExpressionList is called when exiting the statementExpressionList production.
	ExitStatementExpressionList(c *StatementExpressionListContext)

	// ExitEnhancedForStatement is called when exiting the enhancedForStatement production.
	ExitEnhancedForStatement(c *EnhancedForStatementContext)

	// ExitEnhancedForStatementNoShortIf is called when exiting the enhancedForStatementNoShortIf production.
	ExitEnhancedForStatementNoShortIf(c *EnhancedForStatementNoShortIfContext)

	// ExitBreakStatement is called when exiting the breakStatement production.
	ExitBreakStatement(c *BreakStatementContext)

	// ExitContinueStatement is called when exiting the continueStatement production.
	ExitContinueStatement(c *ContinueStatementContext)

	// ExitReturnStatement is called when exiting the returnStatement production.
	ExitReturnStatement(c *ReturnStatementContext)

	// ExitThrowStatement is called when exiting the throwStatement production.
	ExitThrowStatement(c *ThrowStatementContext)

	// ExitSynchronizedStatement is called when exiting the synchronizedStatement production.
	ExitSynchronizedStatement(c *SynchronizedStatementContext)

	// ExitTryStatement is called when exiting the tryStatement production.
	ExitTryStatement(c *TryStatementContext)

	// ExitCatches is called when exiting the catches production.
	ExitCatches(c *CatchesContext)

	// ExitCatchClause is called when exiting the catchClause production.
	ExitCatchClause(c *CatchClauseContext)

	// ExitCatchFormalParameter is called when exiting the catchFormalParameter production.
	ExitCatchFormalParameter(c *CatchFormalParameterContext)

	// ExitCatchType is called when exiting the catchType production.
	ExitCatchType(c *CatchTypeContext)

	// ExitFinallyBlock is called when exiting the finallyBlock production.
	ExitFinallyBlock(c *FinallyBlockContext)

	// ExitTryWithResourcesStatement is called when exiting the tryWithResourcesStatement production.
	ExitTryWithResourcesStatement(c *TryWithResourcesStatementContext)

	// ExitResourceSpecification is called when exiting the resourceSpecification production.
	ExitResourceSpecification(c *ResourceSpecificationContext)

	// ExitResourceList is called when exiting the resourceList production.
	ExitResourceList(c *ResourceListContext)

	// ExitResource is called when exiting the resource production.
	ExitResource(c *ResourceContext)

	// ExitVariableAccess is called when exiting the variableAccess production.
	ExitVariableAccess(c *VariableAccessContext)

	// ExitYieldStatement is called when exiting the yieldStatement production.
	ExitYieldStatement(c *YieldStatementContext)

	// ExitPattern is called when exiting the pattern production.
	ExitPattern(c *PatternContext)

	// ExitTypePattern is called when exiting the typePattern production.
	ExitTypePattern(c *TypePatternContext)

	// ExitExpression is called when exiting the expression production.
	ExitExpression(c *ExpressionContext)

	// ExitPrimary is called when exiting the primary production.
	ExitPrimary(c *PrimaryContext)

	// ExitPrimaryNoNewArray is called when exiting the primaryNoNewArray production.
	ExitPrimaryNoNewArray(c *PrimaryNoNewArrayContext)

	// ExitPNNA is called when exiting the pNNA production.
	ExitPNNA(c *PNNAContext)

	// ExitClassLiteral is called when exiting the classLiteral production.
	ExitClassLiteral(c *ClassLiteralContext)

	// ExitClassInstanceCreationExpression is called when exiting the classInstanceCreationExpression production.
	ExitClassInstanceCreationExpression(c *ClassInstanceCreationExpressionContext)

	// ExitUnqualifiedClassInstanceCreationExpression is called when exiting the unqualifiedClassInstanceCreationExpression production.
	ExitUnqualifiedClassInstanceCreationExpression(c *UnqualifiedClassInstanceCreationExpressionContext)

	// ExitClassOrInterfaceTypeToInstantiate is called when exiting the classOrInterfaceTypeToInstantiate production.
	ExitClassOrInterfaceTypeToInstantiate(c *ClassOrInterfaceTypeToInstantiateContext)

	// ExitTypeArgumentsOrDiamond is called when exiting the typeArgumentsOrDiamond production.
	ExitTypeArgumentsOrDiamond(c *TypeArgumentsOrDiamondContext)

	// ExitArrayCreationExpression is called when exiting the arrayCreationExpression production.
	ExitArrayCreationExpression(c *ArrayCreationExpressionContext)

	// ExitArrayCreationExpressionWithoutInitializer is called when exiting the arrayCreationExpressionWithoutInitializer production.
	ExitArrayCreationExpressionWithoutInitializer(c *ArrayCreationExpressionWithoutInitializerContext)

	// ExitArrayCreationExpressionWithInitializer is called when exiting the arrayCreationExpressionWithInitializer production.
	ExitArrayCreationExpressionWithInitializer(c *ArrayCreationExpressionWithInitializerContext)

	// ExitDimExprs is called when exiting the dimExprs production.
	ExitDimExprs(c *DimExprsContext)

	// ExitDimExpr is called when exiting the dimExpr production.
	ExitDimExpr(c *DimExprContext)

	// ExitArrayAccess is called when exiting the arrayAccess production.
	ExitArrayAccess(c *ArrayAccessContext)

	// ExitFieldAccess is called when exiting the fieldAccess production.
	ExitFieldAccess(c *FieldAccessContext)

	// ExitMethodInvocation is called when exiting the methodInvocation production.
	ExitMethodInvocation(c *MethodInvocationContext)

	// ExitArgumentList is called when exiting the argumentList production.
	ExitArgumentList(c *ArgumentListContext)

	// ExitMethodReference is called when exiting the methodReference production.
	ExitMethodReference(c *MethodReferenceContext)

	// ExitPostfixExpression is called when exiting the postfixExpression production.
	ExitPostfixExpression(c *PostfixExpressionContext)

	// ExitPfE is called when exiting the pfE production.
	ExitPfE(c *PfEContext)

	// ExitPostIncrementExpression is called when exiting the postIncrementExpression production.
	ExitPostIncrementExpression(c *PostIncrementExpressionContext)

	// ExitPostDecrementExpression is called when exiting the postDecrementExpression production.
	ExitPostDecrementExpression(c *PostDecrementExpressionContext)

	// ExitUnaryExpression is called when exiting the unaryExpression production.
	ExitUnaryExpression(c *UnaryExpressionContext)

	// ExitPreIncrementExpression is called when exiting the preIncrementExpression production.
	ExitPreIncrementExpression(c *PreIncrementExpressionContext)

	// ExitPreDecrementExpression is called when exiting the preDecrementExpression production.
	ExitPreDecrementExpression(c *PreDecrementExpressionContext)

	// ExitUnaryExpressionNotPlusMinus is called when exiting the unaryExpressionNotPlusMinus production.
	ExitUnaryExpressionNotPlusMinus(c *UnaryExpressionNotPlusMinusContext)

	// ExitCastExpression is called when exiting the castExpression production.
	ExitCastExpression(c *CastExpressionContext)

	// ExitMultiplicativeExpression is called when exiting the multiplicativeExpression production.
	ExitMultiplicativeExpression(c *MultiplicativeExpressionContext)

	// ExitAdditiveExpression is called when exiting the additiveExpression production.
	ExitAdditiveExpression(c *AdditiveExpressionContext)

	// ExitShiftExpression is called when exiting the shiftExpression production.
	ExitShiftExpression(c *ShiftExpressionContext)

	// ExitRelationalExpression is called when exiting the relationalExpression production.
	ExitRelationalExpression(c *RelationalExpressionContext)

	// ExitEqualityExpression is called when exiting the equalityExpression production.
	ExitEqualityExpression(c *EqualityExpressionContext)

	// ExitAndExpression is called when exiting the andExpression production.
	ExitAndExpression(c *AndExpressionContext)

	// ExitExclusiveOrExpression is called when exiting the exclusiveOrExpression production.
	ExitExclusiveOrExpression(c *ExclusiveOrExpressionContext)

	// ExitInclusiveOrExpression is called when exiting the inclusiveOrExpression production.
	ExitInclusiveOrExpression(c *InclusiveOrExpressionContext)

	// ExitConditionalAndExpression is called when exiting the conditionalAndExpression production.
	ExitConditionalAndExpression(c *ConditionalAndExpressionContext)

	// ExitConditionalOrExpression is called when exiting the conditionalOrExpression production.
	ExitConditionalOrExpression(c *ConditionalOrExpressionContext)

	// ExitConditionalExpression is called when exiting the conditionalExpression production.
	ExitConditionalExpression(c *ConditionalExpressionContext)

	// ExitAssignmentExpression is called when exiting the assignmentExpression production.
	ExitAssignmentExpression(c *AssignmentExpressionContext)

	// ExitAssignment is called when exiting the assignment production.
	ExitAssignment(c *AssignmentContext)

	// ExitLeftHandSide is called when exiting the leftHandSide production.
	ExitLeftHandSide(c *LeftHandSideContext)

	// ExitAssignmentOperator is called when exiting the assignmentOperator production.
	ExitAssignmentOperator(c *AssignmentOperatorContext)

	// ExitLambdaExpression is called when exiting the lambdaExpression production.
	ExitLambdaExpression(c *LambdaExpressionContext)

	// ExitLambdaParameters is called when exiting the lambdaParameters production.
	ExitLambdaParameters(c *LambdaParametersContext)

	// ExitLambdaParameterList is called when exiting the lambdaParameterList production.
	ExitLambdaParameterList(c *LambdaParameterListContext)

	// ExitLambdaParameter is called when exiting the lambdaParameter production.
	ExitLambdaParameter(c *LambdaParameterContext)

	// ExitLambdaParameterType is called when exiting the lambdaParameterType production.
	ExitLambdaParameterType(c *LambdaParameterTypeContext)

	// ExitLambdaBody is called when exiting the lambdaBody production.
	ExitLambdaBody(c *LambdaBodyContext)

	// ExitSwitchExpression is called when exiting the switchExpression production.
	ExitSwitchExpression(c *SwitchExpressionContext)

	// ExitConstantExpression is called when exiting the constantExpression production.
	ExitConstantExpression(c *ConstantExpressionContext)
}
