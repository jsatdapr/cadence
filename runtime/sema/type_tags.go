/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sema

type TypeTag uint64

const NeverTypeTag TypeTag = 0

const (
	UInt8Tag TypeTag = 1 << iota
	UInt16Tag
	UInt32Tag
	UInt64Tag
	UInt128Tag
	UInt256Tag

	Int8Tag
	Int16Tag
	Int32Tag
	Int64Tag
	Int128Tag
	Int256Tag

	Word8Tag
	Word16Tag
	Word32Tag
	Word64Tag

	Fix64Tag
	UFix64Tag

	IntTag
	UIntTag
	StringTag
	CharacterTag
	BoolTag
	NilTag
	VoidTag
	AddressTag
	MetaTag
	AnyStructTag
	AnyResourceTag
	AnyTag

	PathTag
	StoragePathTag
	CapabilityPathTag
	PublicPathTag
	PrivatePathTag

	ArrayTag
	DictionaryTag
	CompositeTag
	ReferenceTag
	ResourceTag

	OptionalTag
	GenericTag
	FunctionTag
	InterfaceTag
	TransactionTag
	RestrictedTag
	CapabilityTag

	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	InvalidTag
)

// Super types
const (
	SignedIntTag = IntTag | Int8Tag | Int16Tag | Int32Tag | Int64Tag | Int128Tag | Int256Tag

	UnsignedIntTag = UIntTag | UInt8Tag | UInt16Tag | UInt32Tag | UInt64Tag | UInt128Tag | UInt256Tag

	IntSuperTypeTag = SignedIntTag | UnsignedIntTag

	AnyStructSuperTypeTag = AnyStructTag | NeverTypeTag | IntSuperTypeTag | StringTag | ArrayTag |
		DictionaryTag | CompositeTag | ReferenceTag | NilTag

	AnyResourceSuperTypeTag = AnyResourceTag | ResourceTag

	AnySuperTypeTag = AnyResourceSuperTypeTag | AnyStructSuperTypeTag
)

// Methods

func CommonSuperType(types ...Type) Type {
	join := NeverTypeTag

	for _, typ := range types {
		join |= typ.Tag()
	}

	return getType(join, types...)
}

func getType(joinedTypeTag TypeTag, types ...Type) Type {
	switch joinedTypeTag {
	case Int8Tag:
		return Int8Type
	case Int16Tag:
		return Int16Type
	case Int32Tag:
		return Int32Type
	case Int64Tag:
		return Int64Type
	case Int128Tag:
		return Int128Type
	case Int256Tag:
		return Int256Type

	case UInt8Tag:
		return UInt8Type
	case UInt16Tag:
		return UInt16Type
	case UInt32Tag:
		return UInt32Type
	case UInt64Tag:
		return UInt64Type
	case UInt128Tag:
		return UInt128Type
	case UInt256Tag:
		return UInt256Type

	case IntTag:
		return IntType
	case UIntTag:
		return UIntType
	case StringTag:
		return StringType
	case NilTag:
		return &OptionalType{
			Type: NeverType,
		}
	case AnyStructTag:
		return AnyStructType
	case AnyResourceTag:
		return AnyResourceType
	case NeverTypeTag:
		return NeverType
	case ArrayTag, DictionaryTag:
		// Contains only arrays or only dictionaries.
		var prevType Type
		for _, typ := range types {
			if prevType == nil {
				prevType = typ
				continue
			}

			if !typ.Equal(prevType) {
				return commonSupertypeOfHeterogeneousTypes(types)
			}
		}

		return prevType

	default:

		// Optional types.
		if joinedTypeTag.containsAny(OptionalTag) {
			// Get the type without the optional flag
			innerTypeTag := joinedTypeTag & (^OptionalTag)
			innerType := getType(innerTypeTag)

			return &OptionalType{
				Type: innerType,
			}
		}

		// Any heterogeneous int subtypes goes here.
		if joinedTypeTag.belongsTo(IntSuperTypeTag) {
			return IntType
		}

		if joinedTypeTag.containsAny(ArrayTag, DictionaryTag) {
			// At this point, the types contains arrays/dictionaries along with other types.
			// So the common supertype could only be AnyStruct, AnyResource or none (both)
			return commonSupertypeOfHeterogeneousTypes(types)
		}

		if joinedTypeTag.belongsTo(AnyStructSuperTypeTag) {
			return AnyStructType
		}

		if joinedTypeTag.belongsTo(AnyResourceSuperTypeTag) {
			return AnyResourceType
		}

		// If nothing works, then there's no common supertype.
		return NeverType
	}
}

func (t TypeTag) containsAny(typeTags ...TypeTag) bool {
	for _, tag := range typeTags {
		if (t & tag) == tag {
			return true
		}
	}

	return false
}

func (t TypeTag) belongsTo(typeTag TypeTag) bool {
	return typeTag.containsAny(t)
}

func commonSupertypeOfHeterogeneousTypes(types []Type) Type {
	var hasStructs, hasResources bool
	for _, typ := range types {
		isResource := typ.IsResourceType()
		hasResources = hasResources || isResource
		hasStructs = hasStructs || !isResource
	}

	if hasResources {
		if hasStructs {
			// If the types has both structs and resources,
			// then there no common super type.
			return NeverType
		}

		return AnyResourceType
	}

	return AnyStructType
}
