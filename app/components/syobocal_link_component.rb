# frozen_string_literal: true

class SyobocalLinkComponent < ApplicationComponent
  def initialize(work:, title: nil)
    @work = work
    @title = title
  end

  def call
    return "-" if work.sc_tid.blank?

    link_to link_title, work.syobocal_url, target: "_blank", rel: "noopener"
  end

  private

  attr_reader :title, :work

  def link_title
    title.presence || work.sc_tid
  end
end
