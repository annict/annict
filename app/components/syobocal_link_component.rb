# frozen_string_literal: true

class SyobocalLinkComponent < ApplicationComponent
  def initialize(work_entity:, title: nil)
    @work_entity = work_entity
    @title = title
  end

  def call
    return "-" if work_entity.syobocal_tid.blank?

    link_to link_title, work_entity.syobocal_url, target: "_blank", rel: "noopener"
  end

  private

  attr_reader :title, :work_entity

  def link_title
    title.presence || work_entity.syobocal_tid
  end
end
