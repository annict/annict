class EditRequestDecorator < ApplicationDecorator
  def status_label
    if opened?
      h.content_tag :span, "オープン", class: "label label-success"
    elsif merged?
      h.content_tag :span, "反映済み", class: "label label-info"
    elsif closed?
      h.content_tag :span, "クローズ", class: "label label-danger"
    end
  end

  def resource_diff_table
    origin_hash = draft_resource.origin.try(:to_diffable_hash).presence || {}
    draft_hash = draft_resource.to_diffable_hash

    diffs = HashDiff.diff(origin_hash, draft_hash)
    origin_values = draft_resource.origin.try(:decorate).try(:to_values)
    draft_values = draft_resource.decorate.to_values

    data = {
      draft_resource: draft_resource,
      diffs: diffs,
      draft_values: draft_values,
      origin_values: origin_values
    }

    h.render("db/application/resource_diff_table", data)
  end
end
