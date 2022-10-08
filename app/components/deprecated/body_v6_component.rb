# frozen_string_literal: true

class Deprecated::BodyV6Component < Deprecated::ApplicationV6Component
  def initialize(view_context, content:, format: :simple, height: nil, class_name: "")
    super view_context
    @content = content
    @format = format
    @height = height
    @class_name = class_name
  end

  def render
    build_html do |h|
      h.tag :div, {
        class: "c-body #{@class_name}",
        data_controller: "body",
        data_body_height: @height
      } do
        h.tag :div, class: "c-body__content", data_body_target: "content" do
          h.html render_content
        end

        h.tag :div,
          class: "c-body__read-more-background d-none w-100",
          data_body_target: "readMoreBackground"

        h.tag :div, {
          class: "c-body__read-more-button d-none text-center w-100",
          data_body_target: "readMoreButton",
          data_action: "click->body#readMore"
        } do
          h.tag :div, class: "c-body__read-more-content small u-fake-link w-100" do
            h.tag :i, class: "fal fa-chevron-double-down me-1"
            h.text t("messages._components.body.view_full_text")
          end
        end
      end
    end
  end

  private

  def render_content
    case @format
    when :simple
      simple_format(@content)
    when :markdown
      view_context.render_markdown(@content)
    end
  end
end
