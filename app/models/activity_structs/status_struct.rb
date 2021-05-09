# frozen_string_literal: true

module ActivityStructs
  class StatusStruct < ActivityStructs::BaseStruct
    attribute :kind, Types::StatusKind
    attribute :anime_id, Types::Integer

    attribute :anime do
      attribute :title, Types::String
      attribute :title_en, Types::String
      attribute :image_path, Types::String
    end
  end
end
