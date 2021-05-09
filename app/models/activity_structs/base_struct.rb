# frozen_string_literal: true

module ActivityStructs
  class BaseStruct < Dry::Struct
    schema schema.strict

    module Types
      include Dry.Types(default: :strict)

      ActivityItemKind = Types::String.enum("record", "status")
      StatusKind = Types::String.enum("plan_to_watch", "watching", "completed", "on_hold", "dropped", "no_status")
    end
  end
end
