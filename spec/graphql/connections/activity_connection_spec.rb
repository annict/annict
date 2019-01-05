# frozen_string_literal: true

describe Connections::ActivityConnection do
  let(:user) { create(:user) }
  let(:id) { GraphQL::Schema::UniqueWithinType.encode(user.class.name, user.id) }
  let!(:activity) { create(:create_episode_record_activity, user: user) }
  let!(:status) { create(:status, user: user) }
  let(:result) do
    query_string = <<~GRAPHQL
      query {
        node(id: "#{id}") {
          id
          ... on User {
            username
            activities(orderBy: { field: CREATED_AT, direction: DESC }) {
              edges {
                node {
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

    res = AnnictSchema.execute(query_string)
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
                node: {
                  state: status.kind.upcase.to_s
                }
              },
              {
                node: {
                  comment: activity.trackable.comment
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
