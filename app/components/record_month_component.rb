# frozen_string_literal: true

class RecordMonthComponent < ApplicationComponent
  def initialize(user_entity:, months:)
    super
    @user_entity = user_entity
    @months = months
  end
end
