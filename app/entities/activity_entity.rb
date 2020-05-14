# frozen_string_literal: true

class ActivityEntity < ApplicationEntity
  attribute? :id, Types::Integer
  attribute? :itemable_type, Types::ActivityResourceKinds
  attribute? :itemable, EpisodeRecordEntity | StatusEntity | WorkRecordEntity
end
