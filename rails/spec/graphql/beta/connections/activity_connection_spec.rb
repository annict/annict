# typed: false
# frozen_string_literal: true

describe Beta::Connections::ActivityConnection do
  let(:user) { create(:user) }
  let(:id) { Beta::AnnictSchema.id_from_object(user, user.class) }
  let!(:activity) { create(:create_episode_record_activity, user: user) }
  let!(:status) { create(:status, user: user) }
  let!(:activity_2) { create(:activity, user: user, itemable: status) }
  let(:result) do
    query_string = <<~GRAPHQL
      query {
        node(id: "#{id}") {
          id
          ... on User {
            username
            activities(orderBy: { field: CREATED_AT, direction: DESC }) {
              edges {
                item {
                  ... on Record {
                    comment
                  }
                  ... on Status {
                    state
                  }
                }
              }
            }
          }
        }
      }
    GRAPHQL

    res = Beta::AnnictSchema.execute(query_string)
    pp(res) if res["errors"]
    res
  end

  it "fetches activities" do
    expected = {
      data: {
        node: {
          id: id,
          username: user.username,
          activities: {
            edges: [
              {
                item: {
                  state: status.kind.upcase.to_s
                }
              },
              {
                item: {
                  comment: activity.itemable.body
                }
              }
            ]
          }
        }
      }
    }
    expect(result.to_h.deep_stringify_keys).to include(expected.deep_stringify_keys)
  end
end
