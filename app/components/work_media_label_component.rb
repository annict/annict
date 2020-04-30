# frozen_string_literal: true

class WorkMediaLabelComponent < ApplicationComponent
  def initialize(work_entity:)
    @work_entity = work_entity
  end

  def call
    content_tag :span, class: "badge u-badge-works" do
      work_entity.media_text
    end
  end

  private

  attr_reader :work_entity
end
