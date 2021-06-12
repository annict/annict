# frozen_string_literal: true

class AnimeRecordEntity < ApplicationEntity
  attribute? :rating_animation, Types::RecordRatingState.optional
  attribute? :rating_music, Types::RecordRatingState.optional
  attribute? :rating_story, Types::RecordRatingState.optional
  attribute? :rating_character, Types::RecordRatingState.optional
  attribute? :rating_overall, Types::RecordRatingState.optional

  def self.from_nodes(work_record_nodes)
    work_record_nodes.map do |work_record_node|
      from_node(work_record_node)
    end
  end

  def self.from_node(work_record_node, user_node: nil)
    attrs = {}

    attrs[:rating_animation] = work_record_node["ratingAnimation"]&.downcase
    attrs[:rating_music] = work_record_node["ratingMusic"]&.downcase
    attrs[:rating_story] = work_record_node["ratingStory"]&.downcase
    attrs[:rating_character] = work_record_node["ratingCharacter"]&.downcase
    attrs[:rating_overall] = work_record_node["ratingOverall"]&.downcase

    new attrs
  end
end
