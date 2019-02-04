pipeline {
    agent any
    stages {
		// At this point the 3 images were created started and stopped, but they exist (docker images)
		// We can create the final artifact, an image with the best of the three		
        stage ('Main image generation') {
            // The final image must be built at the node itsefl, not inside a container
            agent none
            steps {
               // The Dockerfile.artifact copies the code into the image and run the jar generation.
               echo 'Creating the image...'

               // This will search for a Dockerfile.artifact in the working directory and build the image to the local repository
               sh "docker build --no-cache -t \"ditas/vdc-base-image\" -f Dockerfile.artifact ."
               echo "Done"

               echo 'Retrieving Docker Hub password from /opt/ditas-docker-hub.passwd...'
               // Get the password from a file. This reads the file from the host, not the container. Slaves already have the password in there.
               script {
                   password = readFile '/opt/ditas-docker-hub.passwd'
               }
               echo "Done"

               echo 'Login to Docker Hub as ditasgeneric...'
               sh "docker login -u ditasgeneric -p ${password}"
               echo "Done"

               echo "Pushing the image ditas/vdc-base-image:latest..."
               sh "docker push ditas/vdc-base-image:latest"
               echo "Done "
           }
        }
    }
}
