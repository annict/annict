# frozen_string_literal: true

module Canary
  module Types
    module Enums
      class PersonFavoriteOrderField < Canary::Types::Enums::Base
        value "CREATED_AT", "登録日時"
        value "WATCHED_ANIME_COUNT", "見た作品数"
      end
    end
  end
end
