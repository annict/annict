# typed: false
# frozen_string_literal: true

describe "POST /canary/graphql", type: :request do
  let(:work) { create(:work) }
  let(:access_token) { create(:oauth_access_token) }
  let(:id) { Canary::AnnictSchema.id_from_object(work, work.class) }
  let(:query) do
    <<~GRAPHQL
      query($workId: ID!) {
        node(id: $workId) {
          ... on Work {
            id
            databaseId
            title
          }
        }
      }
    GRAPHQL
  end
  let(:headers) { {"Authorization" => "bearer #{access_token.token}"} }

  it "リクエストできること" do
    post "/canary/graphql", params: {variables: {workId: id}, query: query}, headers: headers

    expect(response.status).to eq(200)
    expect(json).to include({
      data: {
        node: {
          databaseId: work.id,
          title: work.title,
          id: id
        }
      }
    }.deep_stringify_keys)
  end
end
