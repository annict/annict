# typed: false
# frozen_string_literal: true

class Deprecated::ErrorPanelV6Component < Deprecated::ApplicationV6Component
  def initialize(view_context, stimulus_controller:)
    super view_context
    @stimulus_controller = stimulus_controller
  end

  def render
    build_html do |h|
      h.tag :div, {
        class: "alert alert-danger c-error-panel d-none",
        "data_#{@stimulus_controller}_target": "errorPanel"
      } do
        h.tag :h4, class: "alert-heading" do
          h.text t("noun.error")
        end

        h.tag :ul, {
          class: "mb-0",
          "data_#{@stimulus_controller}_target": "errorMessageList"
        }
      end
    end
  end
end
