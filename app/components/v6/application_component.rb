# frozen_string_literal: true

module V6
  class ApplicationComponent
    attr_reader :view_context

    delegate :active_link_to, :ann_image_url, :current_user, :display_date, :display_time, :dom_id, :image_tag, :link_to,
      :link_with_domain, :local_url, :page_category, :paginate, :render_markdown, :t, :text_field_tag,
      to: :view_context

    def initialize(view_context)
      @view_context = view_context
    end

    def build_html(&block)
      ActiveSupport::SafeBuffer.new(Htmlrb.build(&block))
    end
  end
end
