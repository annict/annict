# frozen_string_literal: true

module Builder::Activity
  class StatusStruct < Builder::Activity::BaseStruct
    attribute :anime, Types.Instance(Anime)
    attribute :kind, Types::StatusKind
  end
end
