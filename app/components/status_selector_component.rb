# frozen_string_literal: true

class StatusSelectorComponent < ApplicationComponent
  def initialize(work_id:, status_kind: "no_select", page_category:)
    @work_id = work_id
    @status_kind = status_kind
    @page_category = page_category
  end

  def call
    Htmlrb.build do |el|
      el.c_status_selector ":work-id": work_id.to_s, init_status_kind: status_kind, page_category: page_category do
        status_options.each do |(text, value)|
          el.option value: value do
            text
          end
        end
        nil
      end
    end.html_safe
  end

  private

  attr_reader :work_id, :status_kind, :page_category

  def status_options
    Status.kind.options.insert(0, [t("messages.components.status_selector.select_status"), "no_select"])
  end
end
