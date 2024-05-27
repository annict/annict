# typed: false
# frozen_string_literal: true

require "github/markup"

module MarkdownHelper
  def render_markdown(text)
    return "" unless text

    html = GitHub::Markup.render_s(
      GitHub::Markups::MARKUP_MARKDOWN,
      text,
      options: {commonmarker_opts: %i[HARDBREAKS]}
    )
    doc = Nokogiri::HTML::DocumentFragment.parse(html)
    doc.search("img").each do |img|
      img["class"] = "img-fluid img-thumbnail rounded"
    end

    doc.to_html
  end
end
