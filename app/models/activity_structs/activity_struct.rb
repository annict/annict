# frozen_string_literal: true

module ActivityStructs
  class ActivityStruct < ActivityStructs::BaseStruct
    attribute :item_kind, Types::ActivityItemKind
    attribute :outstanding, Types::Bool
    attribute :items_count, Types::Integer
    attribute :created_at, Types::Params::Time

    attribute :items, Types::Array.of(ActivityStructs::StatusStruct)

    attribute :user do
      attribute :username, Types::String
    end
  end
end
