# frozen_string_literal: true

class StatusSelectorComponent < ApplicationComponent
  def initialize(work_id:, page_category:, small: false)
    @work_id = work_id
    @page_category = page_category
    @small = small
  end

  private

  attr_reader :work_id, :page_category, :small

  def status_options
    Status.kind.options.insert(0, [t("messages.components.status_selector.select_status"), "no_select"])
  end
end
