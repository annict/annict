# frozen_string_literal: true

class EpisodeRecordContract < ApplicationContract
  params do
    required(:comment).filled(:stripped_string)
  end
end
