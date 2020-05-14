# frozen_string_literal: true

class ActivityEntity < ApplicationEntity
  attribute? :id, Types::Integer
  attribute? :resource_type, Types::ActivityResourceKinds
  attribute? :resource, EpisodeRecordEntity | StatusEntity | WorkRecordEntity
end
