# typed: false
# frozen_string_literal: true

class SyobocalLinkComponent < ApplicationComponent
  def initialize(work:, title: nil)
    @work = work
    @title = title
  end

  def call
    return "-" if @work.syobocal_tid.blank?

    link_to link_title, @work.syobocal_url, target: "_blank", rel: "noopener"
  end

  private

  def link_title
    @title.presence || @work.syobocal_tid
  end
end
