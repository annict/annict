# frozen_string_literal: true

class ApplicationV6Component
  attr_reader :view_context

  delegate :active_link_to, :v6_ann_image_url, :current_user, :display_date, :display_time, :dom_id, :form_with, :image_tag,
    :link_to, :link_with_domain, :local_url, :page_category, :paginate, :render_markdown, :t, :text_area_tag, :text_field_tag,
    to: :view_context

  def initialize(view_context)
    @view_context = view_context
  end

  def build_html(&block)
    ActiveSupport::SafeBuffer.new(Htmlrb.build(&block))
  end
end
