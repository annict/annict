# frozen_string_literal: true

class ApplicationComponent2
  attr_reader :view_context

  delegate :display_time, :dom_id, :image_tag, :link_to, :paginate, :render_markdown, :t, to: :view_context

  def initialize(view_context)
    @view_context = view_context
  end

  def build_html(&block)
    ActiveSupport::SafeBuffer.new(Htmlrb.build(&block))
  end
end
