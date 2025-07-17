# typed: false
# frozen_string_literal: true

RSpec.describe "POST /graphql", type: :request do
  it "認証済みユーザーがクエリを実行できること" do
    work = create(:work)
    user = create(:user)
    application = create(:oauth_application, owner: user)
    access_token = create(:oauth_access_token, application: application, owner: user)
    id = Beta::AnnictSchema.id_from_object(work, work.class)
    query = <<~GRAPHQL
      query($workId: ID!) {
        node(id: $workId) {
          id
          ... on Work {
            annictId
            title
          }
        }
      }
    GRAPHQL
    headers = {"Authorization" => "bearer #{access_token.token}"}

    post "/graphql", params: {variables: {workId: id}, query: query}, headers: headers

    expect(response.status).to eq(200)
    expect(json).to include({
      data: {
        node: {
          annictId: work.id,
          title: work.title,
          id: id
        }
      }
    }.deep_stringify_keys)
  end

  it "認証なしでクエリを実行した場合401エラーが返ること" do
    work = create(:work)
    id = Beta::AnnictSchema.id_from_object(work, work.class)
    query = <<~GRAPHQL
      query($workId: ID!) {
        node(id: $workId) {
          id
          ... on Work {
            annictId
            title
          }
        }
      }
    GRAPHQL

    post "/graphql", params: {variables: {workId: id}, query: query}

    expect(response.status).to eq(401)
  end

  it "不正なクエリ形式の場合エラーが返ること" do
    user = create(:user)
    application = create(:oauth_application, owner: user)
    access_token = create(:oauth_access_token, application: application, owner: user)
    headers = {"Authorization" => "bearer #{access_token.token}"}
    invalid_query = "invalid query"

    post "/graphql", params: {query: invalid_query}, headers: headers

    expect(response.status).to eq(200)
    expect(json).to have_key("errors")
  end

  it "ミューテーションを実行できること" do
    user = create(:user)
    application = create(:oauth_application, owner: user)
    access_token = create(:oauth_access_token, application: application, owner: user, scopes: "read write")
    work = create(:work)
    episode = create(:episode, work: work)
    mutation = <<~GRAPHQL
      mutation {
        createRecord(input: {episodeId: "#{Beta::AnnictSchema.id_from_object(episode, episode.class)}", ratingState: GREAT}) {
          record {
            id
            ratingState
          }
        }
      }
    GRAPHQL
    headers = {"Authorization" => "bearer #{access_token.token}"}

    post "/graphql", params: {query: mutation}, headers: headers

    expect(response.status).to eq(200)
    expect(json).to have_key("data").or(have_key("errors"))
  end

  it "存在しないノードIDを指定した場合例外が発生すること" do
    user = create(:user)
    application = create(:oauth_application, owner: user)
    access_token = create(:oauth_access_token, application: application, owner: user)
    headers = {"Authorization" => "bearer #{access_token.token}"}
    query = <<~GRAPHQL
      query {
        node(id: "V29yay05OTk5OTk5") {
          id
        }
      }
    GRAPHQL

    expect {
      post "/graphql", params: {query: query}, headers: headers
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
