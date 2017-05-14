# frozen_string_literal: true

Types::WorkImageType = GraphQL::ObjectType.define do
  name "WorkImage"

  implements GraphQL::Relay::Node.interface

  global_id_field :id

  field :internal_url, types.String do
    argument :size, !types.String

    resolve ->(obj, args, ctx) {
      return nil unless ctx[:doorkeeper_token].owner.role.admin?
      ann_image_url obj, :attachment, size: args[:size]
    }
  end
end
