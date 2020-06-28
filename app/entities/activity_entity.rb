# frozen_string_literal: true

class ActivityEntity < ApplicationEntity
  attribute? :database_id, Types::Integer
  attribute? :itemable_type, Types::ActivityResourceKinds
  attribute? :itemable, EpisodeRecordEntity | StatusEntity | WorkRecordEntity
end
