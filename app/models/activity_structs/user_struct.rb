# frozen_string_literal: true

module ActivityStructs
  class UserStruct < ActivityStructs::BaseStruct
    attribute :username, Types::String
    attribute :name, Types::String
    attribute :avatar_path, Types::String
  end
end
