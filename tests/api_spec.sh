# test/api_spec.sh

Describe 'User addition'
    # Test case for successful API request
    Context 'Pass 1'
        It 'returns HTTP 200 OK with expected JSON data'
            When call curl -s -k 'https://localhost:3000/register?user=testuser&token=testpass' 2>&1
            The output should include '"success": true'
            The output should include 'OK: User created.'
        End
    End

    # Test case for failed API request
    Context 'Pass 2'
        It 'returns HTTP 403 Forbidden since user exists'
            When call curl -s -k 'https://localhost:3000/register?user=testuser&token=testpass' 2>&1
            The output should include '"success": false'
            The output should include 'ERR: User already exists.'
        End
    End
End