# frozen_string_literal: true

class EpisodeRecordContract < ApplicationContract
  params do
    required(:rating).maybe(EpisodeRecordEntity::Types::RecordRating)
    required(:comment).maybe(:stripped_string)
    required(:share_to_twitter).filled(:coercible_boolean)
  end
end
