# frozen_string_literal: true

class ApplicationComponent2
  attr_reader :view_context

  delegate :ann_image_url, :current_user, :display_time, :dom_id, :image_tag, :link_to, :paginate, :render_markdown, :t,
    :text_field_tag,
    to: :view_context

  def initialize(view_context)
    @view_context = view_context
  end

  def build_html(&block)
    ActiveSupport::SafeBuffer.new(Htmlrb.build(&block))
  end
end
