# frozen_string_literal: true

class StatusCardComponent < ApplicationComponent
  def initialize(status_entity:)
    @status_entity = status_entity
  end

  private

  attr_reader :status_entity
end
