# frozen_string_literal: true

class AnimeRecordContentComponent < ApplicationComponent
  def initialize(user_entity:, anime_entity:, record_entity:, anime_record_entity:, show_card: true)
    super
    @user_entity = user_entity
    @anime_entity = anime_entity
    @record_entity = record_entity
    @anime_record_entity = anime_record_entity
    @show_card = show_card
  end
end
