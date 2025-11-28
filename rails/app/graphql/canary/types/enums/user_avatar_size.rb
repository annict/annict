# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Enums
      class UserAvatarSize < Canary::Types::Enums::Base
        description "ユーザのアバター画像の大きさ"

        value "size50", "50x50"
        value "size100", "100x100"
        value "size150", "150x150"
        value "size200", "200x200"
      end
    end
  end
end
