# frozen_string_literal: true

class WorkMediaLabelComponent < ApplicationComponent
  inline!

  def initialize(work:)
    @work = work
  end

  def call
    content_tag :span, class: "badge u-badge-works" do
      @work.media_text
    end
  end
end
