# frozen_string_literal: true

class StatusSelectorComponent < ApplicationComponent
  def initialize(work_id:, page_category:, init_kind: "", small: false, class_name: "")
    @anime_id = work_id
    @page_category = page_category
    @init_kind = init_kind
    @small = small
    @class_name = class_name
  end

  private

  def status_selector_class_name
    classes = []
    classes += @class_name.split(" ")
    classes << "c-status-selector--small" if @small
    classes.uniq.join(" ")
  end

  def status_options
    Status.kind.options.insert(0, [t("messages.components.status_selector.select_status"), "no_select"])
  end
end
