# frozen_string_literal: true

class RecordHeaderComponent < ApplicationComponent
  def initialize(user_entity:, record_entity:)
    @user_entity = user_entity
    @record_entity = record_entity
  end

  private

  attr_reader :user_entity, :record_entity
end
